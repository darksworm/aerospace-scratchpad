package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cristianoliveira/aerospace-marks/pkgs/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/aerospacecli"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"go.uber.org/mock/gomock"
)

func TestMoveCmd(t *testing.T) {
	t.Run("fails when missing or empty arguments", func(t *testing.T) {
		command := "move"
		args := []string{command, ""}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if out != "" {
			t.Errorf("Expected empty output, got %s", out)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		snaps.MatchSnapshot(t, cmdAsString, "Output", out, "Error", err.Error())
	})

	t.Run("fails when getting all windows return an erro", func(t *testing.T) {
		command := "move"
		args := []string{command, "test"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
		aerospaceClient.EXPECT().
			GetAllWindows().
			Return(nil, fmt.Errorf("mocked_error")).
			Times(1)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if out != "" {
			t.Errorf("Expected empty output, got %s", out)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		snaps.MatchSnapshot(t, cmdAsString, "Output", out, "Error", err.Error())
	})

	t.Run("moves a window to scratchpad by pattern", func(t *testing.T) {
		command := "move"
		args := []string{command, "MyApp"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		windows := []aerospacecli.Window{
			{
				WindowID: 1,
				AppName:  "MyApp",
			},
			{
				WindowID: 2,
				AppName:  "OtherApp",
			},
		}
		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(windows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(windows[0].WindowID, "scratchpad").
				Return(nil).
				Times(1),

			aerospaceClient.EXPECT().
				SetLayout(windows[0].WindowID, "floating").
				Return(nil).
				Times(1),
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n %+v", err)
		snaps.MatchSnapshot(t, windows, cmdAsString, "Output", out, errorMessage)
	})

	t.Run("fails when moving a window to scratchpad", func(t *testing.T) {
		command := "move"
		args := []string{command, "MyApp"}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		windows := []aerospacecli.Window{
			{
				WindowID: 1,
				AppName:  "MyApp",
			},
			{
				WindowID: 2,
				AppName:  "OtherApp",
			},
		}
		aerospaceClient := aerospacecli_mock.NewMockAeroSpaceClient(ctrl)
		gomock.InOrder(
			aerospaceClient.EXPECT().
				GetAllWindows().
				Return(windows, nil).
				Times(1),

			aerospaceClient.EXPECT().
				MoveWindowToWorkspace(windows[0].WindowID, "scratchpad").
				Return(fmt.Errorf("Window '%+v' already belongs to scratchpad", windows[0])).
				Times(1),

			aerospaceClient.EXPECT().
				SetLayout(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(0),
		)

		cmd := RootCmd(aerospaceClient)
		out, err := testutils.CmdExecute(cmd, args...)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ") + "\n"
		errorMessage := fmt.Sprintf("Error\n%+v", err)
		snaps.MatchSnapshot(t, windows, cmdAsString, "Output", out, errorMessage)
	})
}
