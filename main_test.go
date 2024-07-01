package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCmd(t *testing.T) {
	cmd := newCmd()
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	tmplIn := strings.NewReader(`
{{ range .Commits }}
	{{ .Stats | statsHTMLTable }}
{{end}}
`)
	cmd.SetIn(tmplIn)
	
	cmd.SetArgs([]string{"--dir", "../testtmpldir"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err.Error())
	}
	
	t.Log(buf.String())
}

func TestCommitSlc(t *testing.T) {
	type inputs struct {
		Field string
		Value string
	}
	
	newCommit := func(id, cType string, notes ...note) commit {
		return commit{
			Author:  "author-" + id,
			Message: "message-" + id,
			CC: conventional{
				Notes: notes,
				Scope: "scope-" + id,
				Type:  cType,
			},
		}
	}
	commit1 := newCommit("1", "chore")
	commit2 := newCommit("2", "fix")
	commit3 := newCommit("3", "feat")
	commitWithNotes1 := newCommit("4", "chore", note{Type: "foo", Value: "bar"})
	commitWithNotes2 := newCommit("5", "fix", note{Type: "baz", Value: "fubar"})
	
	commits := commitSlc{commit1, commit2, commit3, commitWithNotes1, commitWithNotes2}
	
	t.Run("KeepByField", func(t *testing.T) {
		tests := []struct {
			name  string
			input inputs
			want  []commit
		}{
			{
				name: "by matching author should pass",
				input: inputs{
					Field: "Author",
					Value: "author-2",
				},
				want: []commit{commit2},
			},
			{
				name: "by matching scope should pass",
				input: inputs{
					Field: "Scope",
					Value: "scope-3",
				},
				want: []commit{commit3},
			},
			{
				name: "by matching type should pass",
				input: inputs{
					Field: "Type",
					Value: "chore",
				},
				want: []commit{commit1, commitWithNotes1},
			},
			{
				name: "without matching value should return empty slice",
				input: inputs{
					Field: "Type",
					Value: "RANDO",
				},
				want: nil,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := commits.KeepByField(tt.input.Field, tt.input.Value)
				mustLen(t, got, len(tt.want))
				for i, want := range tt.want {
					commitEq(t, want, got[i])
				}
			})
		}
	})
	
	t.Run("DropByField", func(t *testing.T) {
		tests := []struct {
			name  string
			input inputs
			want  []commit
		}{
			{
				name: "by matching author should drop",
				input: inputs{
					Field: "Author",
					Value: "author-2",
				},
				want: []commit{commit1, commit3, commitWithNotes1, commitWithNotes2},
			},
			{
				name: "by matching scope should drop",
				input: inputs{
					Field: "Scope",
					Value: "scope-3",
				},
				want: []commit{commit1, commit2, commitWithNotes1, commitWithNotes2},
			},
			{
				name: "by matching type should drop",
				input: inputs{
					Field: "Type",
					Value: "chore",
				},
				want: []commit{commit2, commit3, commitWithNotes2},
			},
			{
				name: "without matching value should return empty slice",
				input: inputs{
					Field: "Type",
					Value: "RANDO",
				},
				want: []commit{commit1, commit2, commit3, commitWithNotes1, commitWithNotes2},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := commits.DropByField(tt.input.Field, tt.input.Value)
				mustLen(t, got, len(tt.want))
				for i, want := range tt.want {
					commitEq(t, want, got[i])
				}
			})
		}
	})
	
	t.Run("KeepByNote", func(t *testing.T) {
		tests := []struct {
			name  string
			input inputs
			want  []commit
		}{
			{
				name: "by matching foo note type should pass",
				input: inputs{
					Field: "foo",
					Value: "bar",
				},
				want: []commit{commitWithNotes1},
			},
			{
				name: "by mismatched foo note type should skip",
				input: inputs{
					Field: "foo",
					Value: "not bar",
				},
				want: nil,
			},
			{
				name: "by matching baz note type should pass",
				input: inputs{
					Field: "baz",
					Value: "fubar",
				},
				want: []commit{commitWithNotes2},
			},
			{
				name: "without matching value should return empty slice",
				input: inputs{
					Field: "RANDO",
					Value: "NOT FOUND",
				},
				want: nil,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := commits.KeepByNote(tt.input.Field, tt.input.Value)
				mustLen(t, got, len(tt.want))
				for i, want := range tt.want {
					commitEq(t, want, got[i])
				}
			})
		}
	})
	
	t.Run("DropByNote", func(t *testing.T) {
		tests := []struct {
			name  string
			input inputs
			want  []commit
		}{
			{
				name: "by matching foo note type should drop",
				input: inputs{
					Field: "foo",
					Value: "bar",
				},
				want: []commit{commit1, commit2, commit3, commitWithNotes2},
			},
			{
				name: "by mismatched foo note type should keep",
				input: inputs{
					Field: "foo",
					Value: "not bar",
				},
				want: []commit{commit1, commit2, commit3, commitWithNotes1, commitWithNotes2},
			},
			{
				name: "by matching baz note type should drop",
				input: inputs{
					Field: "baz",
					Value: "fubar",
				},
				want: []commit{commit1, commit2, commit3, commitWithNotes1},
			},
			{
				name: "without matching value should return empty slice",
				input: inputs{
					Field: "Type",
					Value: "RANDO",
				},
				want: []commit{commit1, commit2, commit3, commitWithNotes1, commitWithNotes2},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := commits.DropByNote(tt.input.Field, tt.input.Value)
				mustLen(t, got, len(tt.want))
				for i, want := range tt.want {
					commitEq(t, want, got[i])
				}
			})
		}
	})
}

func TestNoteSlc_KeepByType(t *testing.T) {
	note1 := note{Type: "foo", Value: "bar"}
	note2 := note{Type: "foo", Value: "baz"}
	note3 := note{Type: "foo", Value: "fubar"}
	note4 := note{Type: "bar", Value: "baz"}
	note5 := note{Type: "baz", Value: "qux"}
	note6 := note{Type: "qux", Value: "quux"}
	
	notes := noteSlc{note1, note2, note3, note4, note5, note6}
	
	got := notes.KeepByType("foo")
	notesEq(t, []note{note1, note2, note3}, got)
	
	got = notes.KeepByType("baz")
	notesEq(t, []note{note5}, got)
}

func Test_statsHTML(t *testing.T) {
	raw := `
 wild-workouts/.gitignore                                             |   36 ++++++++++++++++++++++++++++++++++++
 wild-workouts/LICENSE                                                |   21 +++++++++++++++++++++-----
 wild-workouts/Makefile                                               |   48 ++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/README.md                                              |  107 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/doc.go                                                 |   10 ++++++++++
 wild-workouts/internal/common/auth/http.go                           |   84 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/auth/http_mock.go                      |   44 ++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/auth.go                         |   41 +++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/grpc.go                         |   81 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/net.go                          |   37 +++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/trainer/openapi_client_gen.go   |  568 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/trainer/openapi_types.gen.go    |   57 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/trainings/openapi_client_gen.go | 1004 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/trainings/openapi_types.gen.go  |   60 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/users/openapi_client_gen.go     |  234 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/client/users/openapi_types.gen.go      |   15 +++++++++++++++
 wild-workouts/internal/common/decorator/command.go                   |   27 +++++++++++++++++++++++++++
 wild-workouts/internal/common/decorator/logging.go                   |   56 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/decorator/metrics.go                   |   62 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/decorator/query.go                     |   21 +++++++++++++++++++++
 wild-workouts/internal/common/errors/errors.go                       |   53 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/genproto/trainer/trainer.pb.go         |  315 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/genproto/trainer/trainer_grpc.pb.go    |  209 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/genproto/users/users.pb.go             |  305 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/genproto/users/users_grpc.pb.go        |  137 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/go.mod                                 |   54 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/go.sum                                 |  627 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/logs/cqrs.go                           |   15 +++++++++++++++
 wild-workouts/internal/common/logs/http.go                           |   64 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/logs/logrus.go                         |   31 +++++++++++++++++++++++++++++++
 wild-workouts/internal/common/metrics/dummy.go                       |    7 +++++++
 wild-workouts/internal/common/server/grpc.go                         |   54 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/server/http.go                         |   96 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/server/httperr/http_error.go           |   57 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/tests/clients.go                       |  174 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/tests/e2e_test.go                      |   69 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/tests/hours.go                         |   15 +++++++++++++++
 wild-workouts/internal/common/tests/jwt.go                           |   35 +++++++++++++++++++++++++++++++++++
 wild-workouts/internal/common/tests/wait.go                          |   33 +++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/adapters/hour_firestore_repository.go |  222 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/adapters/hour_memory_repository.go    |   69 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/adapters/hour_mysql_repository.go     |  198 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/adapters/hour_repository_test.go      |  350 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/app/app.go                            |   23 +++++++++++++++++++++++
 wild-workouts/internal/trainer/app/command/cancel_training.go        |   50 ++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/app/command/make_hours_available.go   |   52 ++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/app/command/make_hours_unavailable.go |   52 ++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/app/command/schedule_training.go      |   50 ++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/app/query/hour_availability.go        |   60 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/app/query/types.go                    |    2 ++
 wild-workouts/internal/trainer/domain/hour/availability.go           |   97 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/domain/hour/availability_test.go      |  125 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/domain/hour/hour.go                   |  221 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/domain/hour/hour_test.go              |  248 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/domain/hour/repository.go             |   15 +++++++++++++++
 wild-workouts/internal/trainer/fixtures.go                           |   99 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/go.mod                                |   62 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/go.sum                                |  628 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/main.go                               |   45 +++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/ports/grpc.go                         |   68 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/ports/http.go                         |   49 +++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/ports/openapi_api.gen.go              |  161 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/ports/openapi_types.gen.go            |   57 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/service/application.go                |   51 +++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/internal/trainer/service/component_test.go             |  108 +++++++++++++++++++++++++++++++++++++++++++++++++++++
 wild-workouts/sql/schema.sql                                         |    6 ++++++`
	
	t.Log(statsHTMLTable(raw))
}

func commitEq(t *testing.T, want, got commit) {
	t.Helper()
	
	if want.Author != got.Author {
		t.Errorf("Author do no tmatch:\n\twant: %s\n\tgot: %s", want.Author, got.Author)
	}
	if want.Hash != got.Hash {
		t.Errorf("Hashes do no tmatch:\n\twant: %s\n\tgot: %s", want.Hash, got.Hash)
	}
	if want.HashShort != got.HashShort {
		t.Errorf("HashShorts do no tmatch:\n\twant: %s\n\tgot: %s", want.HashShort, got.HashShort)
	}
	if want.Message != got.Message {
		t.Errorf("Messages do no tmatch:\n\twant: %s\n\tgot: %s", want.Message, got.Message)
	}
	if want.Stats != got.Stats {
		t.Errorf("Stats do not match:\n\twant: %s\n\tgot: %s", want.Stats, got.Stats)
	}
	
	notesEq(t, want.CC.Notes, got.CC.Notes)
}

func notesEq(t *testing.T, want, got []note) {
	t.Helper()
	
	mustLen(t, got, len(want))
	for i, want := range want {
		if got := got[i]; want != got {
			t.Errorf("notes do not match:\n\twant: %#v\n\tgot: %#v", want, got)
		}
	}
}

func mustLen[T any](t *testing.T, slc []T, want int) {
	t.Helper()
	
	if len(slc) != want {
		t.Fatalf("len(slc) = %d, want %d\n\tgot: %v", len(slc), want, slc)
	}
}
