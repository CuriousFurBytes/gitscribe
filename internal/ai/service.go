package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CuriousFurBytes/gitscribe/internal/config"
	"github.com/CuriousFurBytes/gitscribe/internal/execx"
	"github.com/CuriousFurBytes/gitscribe/internal/git"
)

type Request struct {
	Type             string   `json:"type"`
	RepositoryRoot   string   `json:"repository_root"`
	Branch           string   `json:"branch"`
	Diff             string   `json:"diff"`
	Files            []string `json:"files"`
	Style            string   `json:"style"`
	MessageFormat    string   `json:"message_format"`
	PromptComplement string   `json:"prompt_complement"`
	UserFeedback     string   `json:"user_feedback"`
}

type Response struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func Generate(ctx context.Context, repoRoot string, branch string, status git.RepoStatus, cfg config.AIConfig, target string, userFeedback string) (Response, error) {
	if !cfg.Enabled {
		return Response{}, fmt.Errorf("AI generation is disabled")
	}

	diff, err := git.LoadStagedRepositoryDiff(ctx, repoRoot)
	if err != nil {
		return Response{}, err
	}

	files := stagedFiles(status)
	request := Request{
		Type:             target,
		RepositoryRoot:   repoRoot,
		Branch:           branch,
		Diff:             PrepareDiff(diff, cfg.MaxDiffBytes, cfg.PerFileChunkBytes),
		Files:            files,
		Style:            cfg.Style,
		MessageFormat:    cfg.MessageFormat,
		PromptComplement: cfg.PromptComplement,
		UserFeedback:     userFeedback,
	}

	promptText := buildPrompt(request, cfg.Mode)

	expanded := execx.ExpandArgs(cfg.CommandTemplate, map[string]string{
		"{model}":         cfg.Model,
		"{output_format}": cfg.OutputFormat,
	})
	if len(expanded) == 0 {
		return Response{}, fmt.Errorf("AI command template is empty")
	}
	name, args := expanded[0], expanded[1:]

	result, err := execx.Run(ctx, repoRoot, []byte(promptText), name, args...)
	if err != nil {
		return Response{}, fmt.Errorf("AI command failed: %w\n%s", err, result.Output())
	}

	var response Response
	if err := json.Unmarshal([]byte(stripMarkdownFence(unwrapCLIEnvelope(result.Stdout))), &response); err != nil {
		return Response{}, fmt.Errorf("parse AI output: %w\n%s", err, result.Output())
	}
	if strings.TrimSpace(response.Title) == "" {
		return Response{}, fmt.Errorf("AI response title is empty")
	}
	if cfg.Mode == "title_only" {
		response.Body = ""
	}

	return response, nil
}

func unwrapCLIEnvelope(s string) string {
	var envelope struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &envelope); err == nil && envelope.Result != "" && !envelope.IsError {
		return envelope.Result
	}
	return s
}

func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[idx+1:]
	}
	if idx := strings.LastIndex(s, "```"); idx != -1 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

func PrepareDiff(diff string, maxDiffBytes int, perFileChunkBytes int) string {
	if len(diff) <= maxDiffBytes {
		return diff
	}

	segments := strings.Split(diff, "\ndiff --git ")
	var builder strings.Builder
	for i, segment := range segments {
		if i > 0 {
			segment = "diff --git " + segment
		}
		if segment == "" {
			continue
		}
		if len(segment) > perFileChunkBytes {
			segment = segment[:perFileChunkBytes] + "\n...diff truncated...\n"
		}
		if builder.Len()+len(segment) > maxDiffBytes {
			break
		}
		builder.WriteString(segment)
		if !strings.HasSuffix(segment, "\n") {
			builder.WriteString("\n")
		}
	}

	if builder.Len() == 0 {
		return diff[:maxDiffBytes]
	}
	return builder.String()
}

func stagedFiles(status git.RepoStatus) []string {
	files := make([]string, 0, len(status.Files))
	for _, file := range status.Files {
		if file.HasStaged() {
			files = append(files, file.Path)
		}
	}
	return files
}
