package constants

const (
	// DefaultWorkspaceName is the default name of the workspace
	// for the scratchpad
	DefaultScratchpadWorkspaceName = ".scratchpad"
	
	// DefaultGeometry is the default geometry for windows when pulled to current workspace
	// Format: widthPercent%xheightPercent%[@position]
	DefaultGeometry = "60%x90%"
	
	// App launch timeout configuration
	AppLaunchTimeoutSeconds = 3
	AppLaunchMaxRetries     = 3
)

// DefaultScratchpadAppWorkspaces maps common scratchpad application names to their default workspaces
// This helps organize scratchpads by sending each app to its designated workspace when not actively shown
// Note: Only include applications that are typically used as floating scratchpads, not regular tiled apps
var DefaultScratchpadAppWorkspaces = map[string]string{
	"Calculator":         DefaultScratchpadWorkspaceName,
	"Activity Monitor":   DefaultScratchpadWorkspaceName,
	"System Preferences": DefaultScratchpadWorkspaceName,
	"System Settings":    DefaultScratchpadWorkspaceName,
	"Finder":            DefaultScratchpadWorkspaceName,
	"Notes":             DefaultScratchpadWorkspaceName,
	"Stickies":          DefaultScratchpadWorkspaceName,
	"TextEdit":          DefaultScratchpadWorkspaceName,
	"QuickTime Player":  DefaultScratchpadWorkspaceName,
	"Preview":           DefaultScratchpadWorkspaceName,
	"Dictionary":        DefaultScratchpadWorkspaceName,
	"Console":           DefaultScratchpadWorkspaceName,
	"Keychain Access":   DefaultScratchpadWorkspaceName,
	"Digital Color Meter": DefaultScratchpadWorkspaceName,
	"Terminal":          DefaultScratchpadWorkspaceName,
	// Add Arc and Discord as scratchpad apps with their specific workspaces
	"Arc":               "arc",     // Based on your logs, Arc has its own "arc" workspace
	"Discord":           DefaultScratchpadWorkspaceName,  // Discord can go to default scratchpad
	"Thunderbird":       DefaultScratchpadWorkspaceName,  // Thunderbird can go to default scratchpad
	"Mail":              DefaultScratchpadWorkspaceName,  // Apple Mail can go to default scratchpad
	"Linear":            DefaultScratchpadWorkspaceName,  // Linear can go to default scratchpad
	"Spotify":           DefaultScratchpadWorkspaceName,  // Spotify can go to default scratchpad
}

