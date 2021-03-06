package command

import (
	"github.com/hashicorp/serf/cli"
	"github.com/hashicorp/serf/serf"
	"github.com/hashicorp/serf/testutil"
	"strings"
	"testing"
	"time"
)

func TestForceLeaveCommand_implements(t *testing.T) {
	var _ cli.Command = &ForceLeaveCommand{}
}

func TestForceLeaveCommandRun(t *testing.T) {
	a1 := testAgent(t)
	a2 := testAgent(t)
	defer a1.Shutdown()
	defer a2.Shutdown()

	_, err := a1.Join([]string{a2.SerfConfig.MemberlistConfig.BindAddr})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	testutil.Yield()

	// Forcibly shutdown a2 so that it appears "failed" in a1
	if err := a2.Serf().Shutdown(); err != nil {
		t.Fatalf("err: %s", err)
	}

	time.Sleep(a2.SerfConfig.MemberlistConfig.ProbeInterval * 5)

	c := &ForceLeaveCommand{}
	args := []string{
		"-rpc-addr=" + a1.RPCAddr,
		a2.SerfConfig.NodeName,
	}
	ui := new(cli.MockUi)

	code := c.Run(args, ui)
	if code != 0 {
		t.Fatalf("bad: %d. %#v", code, ui.ErrorWriter.String())
	}

	m := a1.Serf().Members()
	if len(m) != 2 {
		t.Fatalf("should have 2 members: %#v", a1.Serf().Members())
	}

	if m[1].Status != serf.StatusLeft {
		t.Fatalf("should be left: %#v", m[1])
	}
}

func TestForceLeaveCommandRun_noAddrs(t *testing.T) {
	c := &ForceLeaveCommand{}
	args := []string{"-rpc-addr=foo"}
	ui := new(cli.MockUi)

	code := c.Run(args, ui)
	if code != 1 {
		t.Fatalf("bad: %d", code)
	}

	if !strings.Contains(ui.ErrorWriter.String(), "node name") {
		t.Fatalf("bad: %#v", ui.ErrorWriter.String())
	}
}
