{pkgs, ...}: {
  packages = with pkgs; [
    go-swag
    protobuf
    buf
    goose
    sqlc
  ];

  languages.go.enable = true;

  scripts.run.exec = ''
    go run cmd/main.go
  '';

  scripts.fmt.exec = ''
    go fmt ./...
  '';

  scripts.tstv.exec = ''
    CLICOLOR_FORCE=1 go test ./... -v
  '';

  scripts.tst.exec = ''
    go test ./...
  '';

  scripts.docs.exec = ''
    swag init -g cmd/main.go
  '';

  scripts.migrate.exec = ''
    goose -dir internal/db/migrations postgres "$DATABASE_URL" up
  '';

  git-hooks.hooks = {
    gotest.enable = true;
    gofmt.enable = true;
    govet.enable = true;
  };

  dotenv.enable = true;
}
