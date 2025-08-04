{pkgs, ...}: {
  packages = with pkgs; [
    # grpc stuff
    grpcurl
    buf

    # sql stuff
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

  scripts.migrate.exec = ''
    goose -dir internal/db/migrations postgres "$DATABASE_URL" up
  '';

  scripts.bump-proto.exec = ''
    git -C proto fetch origin
    git -C proto checkout main
    git -C proto pull --ff-only
    git add proto
    git commit -m "⬆️ bump proto files"
    git push
  '';

  git-hooks.hooks = {
    gotest.enable = true;
    gofmt.enable = true;
    govet.enable = true;
  };

  dotenv.enable = true;
}
