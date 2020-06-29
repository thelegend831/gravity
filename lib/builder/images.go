package builder

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gravitational/gravity/lib/defaults"
	"github.com/gravitational/gravity/lib/docker"
	"github.com/gravitational/gravity/lib/localenv"
	"github.com/gravitational/gravity/lib/pack"

	"github.com/gravitational/trace"
)

// TODO(r0mant): Instead of unpacking provided image (which is slow), see if
// we can modify the build procedure to save a list of embedded images somewhere,
// and fall back to unpacking image for older images.
func GetImages(ctx context.Context, imagePath string) (*InspectResponse, error) {
	env, err := localenv.NewImageEnvironment(imagePath)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer env.Close()
	dir, err := ioutil.TempDir("", "image")
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer os.RemoveAll(dir)
	err = pack.Unpack(env.Packages, env.Manifest.Locator(), dir, nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	images, err := docker.ListImages(ctx, filepath.Join(dir, defaults.RegistryDir))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &InspectResponse{
		Manifest: env.Manifest,
		Images:   images,
	}, nil
}
