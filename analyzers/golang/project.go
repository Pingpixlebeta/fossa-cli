package golang

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/fossas/fossa-cli/analyzers/golang/resolver"
	"github.com/fossas/fossa-cli/files"
)

// A Project is a single folder that forms a coherent "project" for a developer
// and is versioned as a single unit. It may contain multiple Go packages.
type Project struct {
	Tool       resolver.Type // Name of the dependency management tool used by the project, if any.
	Manifest   string        // Absolute path to the tool's manifest file for this project, if any.
	Dir        string        // Absolute path of the first-party code folder, if any.
	ImportPath string        // Import path prefix of project code.
}

// Project calculates the project containing any Go package.
//
// This function searches upwards from the Go package's directory, looking for
// lockfiles of supported dependency management tools. If none are found, it
// fails.
//
// The rationale for this design is that the packages in a "project" are
// versioned together. There are two reasonable ways to capture the notion of
// "versioned together":
//
//   1. The nearest lockfile. The nearest lockfile to the package probably locks
//      the dependencies of the package.
//   2. The nearest VCS repository. The nearest VCS repository probably contains
//      the current "project" being worked on. The only common exception to this
//      is monorepos, in which case all the contents of the repository are
//      probably internal, so allowing packages within the repository to be
//      unresolved is probably acceptable.
//
// There are a couple issues with both of these:
//
//   1. The nearest lockfile is not guaranteed to exist. When it does, it's not
//      guaranteed to be _the_ semantic lockfile for the package -- this is
//      merely a very common convention, not a requirement.
//   2. The package is not guaranteed to be in a VCS repository. When it is, the
//      repository might be extremely weird. One example of this is a repository
//      containing the entire $GOPATH (which is a reasonable convention that
//      some early adopters of Go used).
//
// This function tries its best to mitigate both of these issues:
//
//   1. The nearest lockfile is used for resolving versions. This is a very
//      strong convention.
//   2. The nearest VCS repository is used for determining allowed unresolved
//      import paths. This is also a very strong convention.
//
// Both of these assumptions can be overridden by the user.
func (a *Analyzer) Project(pkg string) (Project, error) {
	// Check for a cached project.
	cached, ok := a.projectCache[pkg]
	if ok {
		return cached, nil
	}

	// Get the package directory.
	dir, err := a.Dir(pkg)
	if err != nil {
		return Project{}, err
	}

	// Find the nearest lockfile.
	tool, manifestDir, err := NearestLockfile(dir)

	// Find the nearest VCS repository.
	_, repoRoot, err := NearestVCS(dir)
	if err != nil {
		return Project{}, err
	}

	// Handle nested vendor folders: in order to do lockfile computations, the
	// "project" of a nested dependency is its parent.
	parent := VendorParent(dir)

	// Project root is the lower of the nearest VCS or the vendor parent.
	projectDir := repoRoot
	if strings.HasPrefix(parent, repoRoot) {
		projectDir = parent
	}

	// Compute the project import path prefix.
	if os.Getenv("GOPATH") == "" {
		return Project{}, errors.New("no $GOPATH set")
	}
	gopath, err := filepath.Abs(os.Getenv("GOPATH"))
	if err != nil {
		return Project{}, errors.Wrap(err, "could not get absolute $GOPATH")
	}
	importPrefix, err := filepath.Rel(filepath.Join(gopath, "src"), projectDir)
	if err != nil {
		return Project{}, errors.Wrap(err, "could not compute import prefix")
	}

	// Cache the computed project.
	project := Project{
		Tool:       tool,
		Manifest:   manifestDir,
		Dir:        projectDir,
		ImportPath: importPrefix,
	}
	a.projectCache[pkg] = project
	return project, nil
}

// This is a monomorphic Either monad. I miss Haskell.
type eitherStr struct {
	result string
	err    error
}

func (r *eitherStr) Bind(tool string, find func(pathElems ...string) (bool, error), pathElems ...string) {
	if r.err != nil {
		return
	}

	ok, err := find(pathElems...)
	if err != nil {
		r.err = err
	}
	if ok {
		r.result = tool
	}
}

func (r *eitherStr) Find(tool string, pathElems ...string) {
	r.Bind(tool, files.Exists, pathElems...)
}

func (r *eitherStr) FindFolder(tool string, pathElems ...string) {
	r.Bind(tool, files.ExistsFolder, pathElems...)
}
