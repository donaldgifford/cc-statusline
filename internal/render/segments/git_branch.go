package segments

import (
	"github.com/donaldgifford/cc-statusline/internal/git"
	"github.com/donaldgifford/cc-statusline/internal/model"
	"github.com/donaldgifford/cc-statusline/internal/render/theme"
)

// GitBranch displays the current git branch wrapped in parentheses.
type GitBranch struct{}

// Name implements render.Segment.
func (GitBranch) Name() string { return "git_branch" }

// Source implements render.Segment.
func (GitBranch) Source() string { return SourceStable }

// Render implements render.Segment.
func (GitBranch) Render(data *model.StatusData, th *theme.Theme) (string, error) {
	branch := git.Branch(data.CWD)
	if branch == "" {
		return "", nil
	}
	return th.Colorize("git_branch", "("+branch+")"), nil
}
