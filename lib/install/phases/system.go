package phases

import (
	"context"

	"github.com/gravitational/gravity/lib/constants"
	"github.com/gravitational/gravity/lib/fsm"
	"github.com/gravitational/gravity/lib/ops"
	"github.com/gravitational/gravity/lib/utils"

	"github.com/gravitational/configure"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
)

// NewSystem returns a new "system" phase executor
func NewSystem(p fsm.ExecutorParams, operator ops.Operator, remote fsm.Remote) (*systemExecutor, error) {
	logger := &fsm.Logger{
		FieldLogger: logrus.WithFields(logrus.Fields{
			constants.FieldInstallPhase: p.Phase.ID,
			constants.FieldAdvertiseIP:  p.Phase.Data.Server.AdvertiseIP,
			constants.FieldHostname:     p.Phase.Data.Server.Hostname,
		}),
		Key:      opKey(p.Plan),
		Operator: operator,
		Server:   p.Phase.Data.Server,
	}
	return &systemExecutor{
		FieldLogger:    logger,
		ExecutorParams: p,
		remote:         remote,
	}, nil
}

type systemExecutor struct {
	// FieldLogger is used for logging
	logrus.FieldLogger
	// ExecutorParams is common executor params
	fsm.ExecutorParams
	// remote specifies the server remote control interface
	remote fsm.Remote
}

// Execute executes the system phase
func (p *systemExecutor) Execute(ctx context.Context) error {
	locator := p.Phase.Data.Package
	p.Progress.NextStep("Installing system service %v:%v",
		locator.Name, locator.Version)
	p.Infof("Installing system service %v:%v", locator.Name, locator.Version)
	args := []string{"--debug", "system", "reinstall", locator.String()}
	if len(p.Phase.Data.Labels) != 0 {
		labels := configure.KeyVal(p.Phase.Data.Labels)
		args = append(args, "--labels", labels.String())
	}
	out, err := utils.RunGravityCommand(ctx, p.FieldLogger, args...)
	return trace.Wrap(err, "failed to install system service: %s", string(out))
}

// Rollback is no-op for this phase
func (*systemExecutor) Rollback(ctx context.Context) error {
	return nil
}

// PreCheck makes sure this phase is executed on a proper node
func (p *systemExecutor) PreCheck(ctx context.Context) error {
	err := p.remote.CheckServer(ctx, *p.Phase.Data.Server)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// PostCheck is no-op for this phase
func (*systemExecutor) PostCheck(ctx context.Context) error {
	return nil
}