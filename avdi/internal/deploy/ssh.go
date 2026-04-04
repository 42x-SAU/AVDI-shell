package deploy

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Options configures remote Docker-based agent deployment over SSH.
type Options struct {
	Host   string
	Port   int
	User   string
	Secret string // password (SSH key auth will be added later)

	// ServerURL is the AVDI HTTP API base URL as reachable from the agent host (e.g. http://10.0.0.1:8081).
	ServerURL string
	AgentName string

	Image         string
	ContainerName string
	SkipPull      bool
}

// DeployDockerAgent connects with password authentication, ensures Docker is available,
// optionally pulls the image, and runs the agent container with SERVER_URL and AGENT_NAME.
func DeployDockerAgent(opts Options) error {
	if err := validate(opts); err != nil {
		return err
	}
	if opts.Port == 0 {
		opts.Port = 22
	}
	if opts.ContainerName == "" {
		opts.ContainerName = "avdi-agent"
	}

	addr := net.JoinHostPort(opts.Host, fmt.Sprintf("%d", opts.Port))
	cfg := &ssh.ClientConfig{
		User: opts.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(opts.Secret),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         45 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("ssh session: %w", err)
	}
	defer session.Close()

	script := buildRemoteScript(opts)
	var stdout, stderr bytes.Buffer
	session.Stdin = strings.NewReader(script)
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run("/bin/sh -s"); err != nil {
		return fmt.Errorf("remote script failed: %w\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	out := strings.TrimSpace(stdout.String() + stderr.String())
	if out != "" {
		fmt.Println(out)
	}
	return nil
}

func validate(o Options) error {
	if o.Host == "" {
		return fmt.Errorf("ssh host is required")
	}
	if o.User == "" {
		return fmt.Errorf("ssh user is required")
	}
	if o.Secret == "" {
		return fmt.Errorf("ssh password is empty")
	}
	if o.ServerURL == "" {
		return fmt.Errorf("server URL is required (AVDI API reachable from the agent machine)")
	}
	if o.AgentName == "" {
		return fmt.Errorf("agent name is required")
	}
	if o.Image == "" {
		return fmt.Errorf("docker image is required")
	}
	return nil
}

func buildRemoteScript(o Options) string {
	var b strings.Builder
	q := shellQuote
	fmt.Fprintf(&b, "set -e\n")
	fmt.Fprintf(&b, "IMAGE=%s\n", q(o.Image))
	fmt.Fprintf(&b, "SERVER_URL=%s\n", q(o.ServerURL))
	fmt.Fprintf(&b, "AGENT_NAME=%s\n", q(o.AgentName))
	fmt.Fprintf(&b, "CONTAINER=%s\n", q(o.ContainerName))
	fmt.Fprintf(&b, `if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found on remote host" >&2
  exit 1
fi
`)
	if !o.SkipPull {
		fmt.Fprintf(&b, `docker pull "$IMAGE" || {
  echo "docker pull failed for: $IMAGE" >&2
  echo "Typical fixes:" >&2
  echo "  - Use a real image (build, tag, push to Docker Hub / GHCR / your registry)." >&2
  echo "  - On this host: docker login <registry>, then redeploy." >&2
  echo "  - If the image is already here (docker load / local build): redeploy with --skip-pull." >&2
  exit 1
}
`)
	}
	if o.SkipPull {
		fmt.Fprintf(&b, `if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "image not found locally: $IMAGE (use docker load / docker build, or omit --skip-pull to pull)" >&2
  exit 1
fi
`)
	}
	fmt.Fprintf(&b, `docker rm -f "$CONTAINER" 2>/dev/null || true
docker run -d --name "$CONTAINER" --restart unless-stopped \
  -e SERVER_URL="$SERVER_URL" \
  -e AGENT_NAME="$AGENT_NAME" \
  "$IMAGE"
echo "container started: $CONTAINER"
`)
	return b.String()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
