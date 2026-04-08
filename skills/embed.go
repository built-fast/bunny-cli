package skills

import "embed"

//go:embed bunny
var Content embed.FS

// SkillMD returns the raw contents of the embedded SKILL.md file.
func SkillMD() ([]byte, error) {
	return Content.ReadFile("bunny/SKILL.md")
}
