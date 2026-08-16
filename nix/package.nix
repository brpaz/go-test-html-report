{ lib, buildGoModule }:

buildGoModule {
  pname = "go-test-html-report";
  version = "0.1.3";
  src = ../.;
  subPackages = [ "cmd/go-test-html-report" ];
  vendorHash = "sha256-Ld0nVtjHTcws03etLguxsxzfEKzod8LdGMHbFEjvUws=";
  ldflags = [
    "-s"
    "-w"
    "-X main.version=0.1.3"
  ];

  meta = with lib; {
    description = "A Golang library for generating HTML reports from go test results";
    homepage = "https://github.com/brpaz/go-test-html-report";
    license = licenses.mit;
    mainProgram = "go-test-html-report";
  };
}
