package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/goccy/go-yaml"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
)

type SnapshotWorkspace struct {
	Workspace       string `yaml:"workspace"`
	FocusedWindowID int    `yaml:"focused-window-id,omitempty"`
}

type SnapshotWindow struct {
	WindowID                    int    `yaml:"window-id"`
	WindowTitle                 string `yaml:"window-title,omitempty"`
	WindowLayout                string `yaml:"window-layout,omitempty"`
	WindowParentContainerLayout string `yaml:"parent-layout,omitempty"`
	AppName                     string `yaml:"app-name,omitempty"`
	AppBundleID                 string `yaml:"app-bundle-id,omitempty"`
	Workspace                   string `yaml:"workspace,omitempty"`
}

type SnapshotContext struct {
	Workspaces []SnapshotWorkspace `yaml:"workspaces,omitempty"`
	Windows    []SnapshotWindow    `yaml:"windows,omitempty"`
	Raw        any                 `yaml:"raw,omitempty"`
}

const (
	contextIndent = 2
	blockIndent   = 4
)

// MatchSnapshot formats a human-readable snapshot with Context, Command, and Output sections.
// It accepts both the legacy snapshot arguments and the new compact signature.
// Examples:
//
//	MatchSnapshot(t, tree, cmd, "Output", out, err)
//	MatchSnapshot(t, cmd, "Output", out, err)
//	MatchSnapshot(t, tree, cmd, out, err)
func MatchSnapshot(t testing.TB, values ...any) {
	t.Helper()
	ctx, command, stdout, errVal := normalizeSnapshotValues(values...)
	snaps.MatchSnapshot(t, HumanReadableSnapshot(ctx, command, stdout, errVal))
}

// HumanReadableSnapshot builds the formatted snapshot content.
func HumanReadableSnapshot(ctx any, command string, stdout string, errVal any) string {
	snapContext := buildSnapshotContext(ctx)

	var b strings.Builder

	// Context
	b.WriteString("Context:\n")
	b.WriteString(indent(renderContext(snapContext), contextIndent))

	// Command
	trimmedCmd := strings.TrimSpace(command)
	if trimmedCmd == "" {
		trimmedCmd = "(none)"
	}
	b.WriteString("Command: |\n")
	b.WriteString(indent("$ "+trimmedCmd+"\n", contextIndent))

	// Output
	status := "success"
	errorText := normalizeError(errVal)
	if errorText != "" {
		status = "error"
	}

	b.WriteString("Output:\n")
	b.WriteString(indent("status: "+status+"\n", contextIndent))

	if stdout != "" {
		b.WriteString(indent("stdout: |\n", contextIndent))
		b.WriteString(indentBlock(stdout, blockIndent))
	} else {
		b.WriteString(indent("stdout: \"\"\n", contextIndent))
	}

	if errorText != "" {
		b.WriteString(indent("error: |\n", contextIndent))
		b.WriteString(indentBlock(errorText, blockIndent))
	} else {
		b.WriteString(indent("error: \"\"\n", contextIndent))
	}

	return b.String()
}

func normalizeSnapshotValues(values ...any) (any, string, string, any) {
	var ctx any
	var command string
	var stdout string
	var errVal any

	for _, value := range values {
		switch v := value.(type) {
		case string:
			ctx, command, stdout, errVal = handleStringArg(ctx, command, stdout, errVal, v)
		case error:
			errVal = firstNonNilError(errVal, v)
		default:
			if ctx == nil {
				ctx = value
			}
		}
	}

	return ctx, strings.TrimSpace(command), stdout, errVal
}

func handleStringArg(
	ctx any,
	command, stdout string,
	errVal any,
	value string,
) (any, string, string, any) {
	if isLabel(value) {
		return ctx, command, stdout, errVal
	}
	if command == "" {
		return ctx, value, stdout, errVal
	}
	if stdout == "" {
		return ctx, command, value, errVal
	}
	if errVal == nil {
		return ctx, command, stdout, value
	}

	return ctx, command, stdout, errVal
}

func firstNonNilError(current any, candidate error) any {
	if current != nil || candidate == nil {
		return current
	}
	return candidate
}

func isLabel(value string) bool {
	switch value {
	case "Output", "Error":
		return true
	default:
		return false
	}
}

func normalizeError(errVal any) string {
	switch v := errVal.(type) {
	case nil:
		return ""
	case error:
		if v == nil {
			return ""
		}
		return v.Error()
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || trimmed == "<nil>" || trimmed == "Error\n<nil>" {
			return ""
		}
		return trimmed
	default:
		return fmt.Sprint(v)
	}
}

func buildSnapshotContext(ctx any) SnapshotContext {
	switch v := ctx.(type) {
	case nil:
		return SnapshotContext{}
	case []AeroSpaceTree:
		return contextFromTree(v)
	case []windows.Window:
		return contextFromWindows(v)
	default:
		return SnapshotContext{Raw: v}
	}
}

func contextFromTree(tree []AeroSpaceTree) SnapshotContext {
	var workspaces []SnapshotWorkspace
	var windows []SnapshotWindow

	for _, node := range tree {
		if node.Workspace != nil {
			workspaces = append(workspaces, SnapshotWorkspace{
				Workspace:       node.Workspace.Workspace,
				FocusedWindowID: node.FocusedWindowID,
			})
		}
		windows = append(windows, contextFromWindows(node.Windows).Windows...)
	}

	return SnapshotContext{Workspaces: workspaces, Windows: windows}
}

func contextFromWindows(win []windows.Window) SnapshotContext {
	var windows []SnapshotWindow
	for _, w := range win {
		windows = append(windows, SnapshotWindow{
			WindowID:                    w.WindowID,
			WindowTitle:                 w.WindowTitle,
			WindowLayout:                w.WindowLayout,
			WindowParentContainerLayout: w.WindowParentContainerLayout,
			AppName:                     w.AppName,
			AppBundleID:                 w.AppBundleID,
			Workspace:                   w.Workspace,
		})
	}
	return SnapshotContext{Windows: windows}
}

func renderContext(ctx SnapshotContext) string {
	marshaled, err := yaml.Marshal(ctx)
	if err != nil {
		return "{}\n"
	}
	return string(marshaled)
}

func indent(text string, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	var b strings.Builder
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "" && i == len(lines)-1 {
			continue
		}
		b.WriteString(pad)
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func indentBlock(text string, spaces int) string {
	trimmed := strings.TrimRight(text, "\n")
	if trimmed == "" {
		return indent("\n", spaces)
	}
	return indent(trimmed+"\n", spaces)
}
