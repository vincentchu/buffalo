package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/genny/v2"
	"github.com/gobuffalo/packr/v2/jam"
)

// Cleanup all of the generated files
func Cleanup(opts *Options) genny.RunFn {
	return func(r *genny.Runner) error {
		defer os.RemoveAll(filepath.Join(opts.Root, "a"))
		var err error
		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(opts.Root)
		if err == nil {
			log.Println(opts.Root, "=", fileInfo.Mode())
		}
		if err := jam.Clean(); err != nil {
			return err
		}

		opts.rollback.Range(func(k, v interface{}) bool {
			f := genny.NewFileS(k.(string), v.(string))
			r.Logger.Debugf("Rollback: %s", f.Name())
			fileInfo, err = os.Stat(k.(string))
			if err == nil {
				log.Println(fileInfo)
				log.Println(fileInfo.Mode())
			}
			if err = r.File(f); err != nil {
				fmt.Printf("cleanup error: %s - %s", f.Name(), err)
				return false
			}
			r.Disk.Remove(f.Name())
			return true
		})
		if err != nil {
			return err
		}
		for _, f := range r.Disk.Files() {
			if _, keep := opts.keep.Load(f.Name()); keep {
				// Keep this file
				continue
			}
			r.Disk.Delete(f.Name())
		}
		if envy.Mods() && opts.WithBuildDeps {
			if err := r.Exec(exec.Command("go", "mod", "tidy")); err != nil {
				return err
			}
		}
		return nil
	}
}
