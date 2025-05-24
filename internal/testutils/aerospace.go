package testutils

import "github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"

type AeroSpaceTree struct {
	Windows []aerospacecli.Window
	Workspace *aerospacecli.Workspace
	FocusedWindowId int
}

func ExtractAllWindows(tree []AeroSpaceTree) []aerospacecli.Window {
	var allWindows []aerospacecli.Window
	for _, t := range tree {
		allWindows = append(allWindows, t.Windows...)
	}
	return allWindows
}

func ExtractFocusedTree(tree []AeroSpaceTree) *AeroSpaceTree {
	for _, t := range tree {
		if t.FocusedWindowId != 0 {
			return &t
		}
	}
	return nil
}

func ExtractFocusedWindow(tree []AeroSpaceTree) *aerospacecli.Window {
	for _, t := range tree {
		if t.FocusedWindowId != 0 {
			for _, window := range t.Windows {
				if window.WindowID == t.FocusedWindowId {
					return &window
				}
			}
		}
	}
	return nil
}

