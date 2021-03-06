// Guess project "type" from the files present.
package detect

import (
	"errors"
	"path"

	"github.com/heyLu/qst/fileutil"
)

type Project struct {
	Id       string
	Commands Commands
	Detect   Matcher
}

type Matcher func(string) bool

type Commands map[string]string

var ProjectTypes = []*Project{
	&Project{"c", Commands{"run": "gcc -o $(basename {file} .c) {file} && ./$(basename {file} .c)"},
		matchPattern("*.c")},
	&Project{"clojure/leiningen", Commands{"build": "lein uberjar", "run": "lein run", "test": "lein test"},
		matchFile("project.clj")},
	&Project{"coffeescript", Commands{"run": "coffee {file}"}, matchPattern("*.coffee")},
	&Project{"docker/fig", Commands{"build": "fig build", "run": "fig up"}, matchFile("fig.yml")},
	&Project{"docker", Commands{"build": "docker build ."}, matchFile("Dockerfile")},
	&Project{"executable", Commands{"run": "{file}"}, executable},
	&Project{"go", Commands{"build": "go build {file}", "run": "go build $(basename {file}) && ./$(basename {file} .go)",
		"test": "go test"}, matchPattern("*.go")},
	&Project{"haskell/cabal", Commands{"build": "cabal sandbox init && cabal install --only-dependencies && cabal build",
		"run": "cabal sandox init && cabal run", "test": "cabal sandbox init && cabal test"}, matchPattern("*.cabal")},
	&Project{"haskell", Commands{"run": "runhaskell {file}"}, haskell},
	&Project{"idris", Commands{"run": "idris -o $(basename {file} .idr) {file} && ./$(basename {file} .idr)"},
		matchPattern("*.idr")},
	&Project{"java/maven", Commands{"build": "mvn compile", "test": "mvn compile test"}, matchFile("pom.xml")},
	&Project{"javascript/npm", Commands{"build": "npm install", "run": "npm start", "test": "npm test"},
		matchFile("package.json")},
	&Project{"javascript/meteor", Commands{"run": "meteor"}, matchFile(".meteor/.id")},
	&Project{"javascript", Commands{"run": "node {file}"}, matchPattern("*.js")},
	&Project{"jekyll", Commands{"build": "jekyll build", "run": "jekyll serve --watch"}, jekyll},
	&Project{"julia", Commands{"run": "julia {file}"}, matchPattern("*.jl")},
	&Project{"latex", Commands{"run": "pdflatex {file}"}, latex},
	&Project{"python/django", Commands{"build": "python manage.py syncdb", "run": "python manage.py runserver",
		"test": "python manage.py test"}, matchFile("manage.py")},
	&Project{"python", Commands{"run": "python {file}"}, matchPattern("*.py")},
	&Project{"ruby/rails", Commands{"build": "bundle exec rake db:migrate", "run": "rails server",
		"test": "bundle exec rake test"}, matchFile("bin/rails")},
	&Project{"ruby/rake", Commands{"run": "rake", "test": "rake test"}, matchFile("Rakefile")},
	&Project{"ruby", Commands{"run": "ruby {file}"}, matchPattern("*.rb")},
	&Project{"rust/cargo", Commands{"build": "cargo build", "run": "cargo run", "test": "cargo test"},
		matchFile("Cargo.toml")},
	&Project{"rust", Commands{"run": "rustc {file} && ./$(basename {file} .rs)"}, matchPattern("*.rs")},
	&Project{"cmake", Commands{"build": "mkdir .build && cd .build && cmake .. && make"}, matchFile("CMakeLists.txt")},
	&Project{"make", Commands{"run": "make", "test": "make test"}, matchFile("Makefile")},
	&Project{"procfile", Commands{"run": "$(sed -n 's/^web: //p' Procfile)"}, matchFile("Procfile")},
}

func Detect(file string) (*Project, error) {
	for _, project := range ProjectTypes {
		if project.Detect(file) {
			return project, nil
		}
	}

	return nil, errors.New("no project type matches")
}

func DetectAll(file string) []*Project {
	projects := make([]*Project, 0, len(ProjectTypes))

	for _, project := range ProjectTypes {
		if project.Detect(file) {
			n := len(projects)
			projects = projects[0 : n+1]
			projects[n] = project
		}
	}

	return projects
}

func GetById(id string) *Project {
	for _, project := range ProjectTypes {
		if project.Id == id {
			return project
		}
	}
	return nil
}

func matchingFileOrDir(file string, pattern string) bool {
	if fileutil.IsFile(file) {
		_, f := path.Split(file)
		isMatch, _ := path.Match(pattern, f)
		return isMatch
	} else {
		return fileutil.MatchExists(path.Join(file, pattern))
	}
}

func hasFile(fileOrDir string, file string) bool {
	return fileutil.IsFile(fileutil.Join(fileOrDir, file))
}

func matchPattern(ext string) Matcher {
	return func(file string) bool {
		return matchingFileOrDir(file, ext)
	}
}

func matchFile(fileName string) Matcher {
	return func(file string) bool {
		return hasFile(file, fileName)
	}
}

func executable(file string) bool {
	return fileutil.IsExecutable(file)
}

func haskell(file string) bool {
	return matchingFileOrDir(file, "*.hs") || matchingFileOrDir(file, "*.lhs")
}

func jekyll(file string) bool {
	return hasFile(file, "_config.yml") || fileutil.IsDir(fileutil.Join(file, "_posts"))
}

func latex(file string) bool {
	return matchingFileOrDir(file, "*.latex") || matchingFileOrDir(file, "*.tex")
}
