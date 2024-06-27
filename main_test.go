package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestCmd(t *testing.T) {
	cmd := newCmd()
	
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	
	tmplIn := strings.NewReader(`
{{ range .Commits }}
	{{ .Stats | statsHTML }}
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
	
	for _, l := range strings.Split(statsHTML(raw), "</br>") {
		fmt.Println(l)
	}
}
