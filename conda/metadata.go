package conda

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type indexJson struct {
	Name    string   `json:"name"`
	Entry   string   `json:"app_entry"`
	Version string   `json:"version"`
	Build   string   `json:"build"`
	License string   `json:"license"`
	Subdir  string   `json:"subdir"`
	Depends []string `json:"depends"`
}

func (c *indexJson) Load(dir string, extraDeps map[string][]string) error {
	if b, err := os.ReadFile(path.Join(dir, "info", "index.json")); err != nil {
		return err
	} else if err := json.Unmarshal(b, c); err != nil {
		return err
	}
	for _, depStr := range extraDeps[c.Name] {
		fmt.Printf("Adding extra dependency %s to %s\n",
			depStr, c.Name)
		dep := depStr
		found := false
		for _, d := range c.Depends {
			if d == dep {
				fmt.Printf("Dependency already present.")
				found = true
				break
			}
		}
		if !found {
			c.Depends = append(c.Depends, dep)
		}
	}
	return nil
}
