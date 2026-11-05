package skills

import (
	"fmt"
	"os"
)

// zipSearchDepth allows SKILL.md at the archive root or inside the single
// folder archivers usually wrap the content in.
const zipSearchDepth = 1

// InstallFromZip extracts the .zip at zipPath, which must contain one skill
// folder (SKILL.md at the root or one level down), and installs it into
// userDir/<name> as InstallFromFolder does.
func InstallFromZip(zipPath, userDir string, overwrite bool) (Skill, error) {
	tmp, err := os.MkdirTemp("", "gocode-skill-zip-")
	if err != nil {
		return Skill{}, err
	}
	defer os.RemoveAll(tmp) // best effort: only a temporary copy remains
	if err := extractZip(zipPath, tmp); err != nil {
		return Skill{}, err
	}
	dirs, err := findSkillDirs(tmp, zipSearchDepth)
	if err != nil {
		return Skill{}, fmt.Errorf("%s: %w", zipPath, err)
	}
	if len(dirs) > 1 {
		return Skill{}, fmt.Errorf("%s contains %d skills; expected one", zipPath, len(dirs))
	}
	return InstallFromFolder(dirs[0], userDir, overwrite)
}
