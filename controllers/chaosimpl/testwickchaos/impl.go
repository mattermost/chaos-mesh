package testwickchaos

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

type Impl struct {
	client.Client
	Log     logr.Logger
	decoder *utils.ContianerRecordDecoder
}

// Apply applies TestWickChaos
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("Initiating TestWick experiment")
	testWick := obj.(*v1alpha1.TestWickChaos)
	impl.Log.Info("Creating installations", "installations", testWick.TestWick.Samples)

	arguments := []string{
		"test",
		"--channel-messages", *testWick.TestWick.ChannelMessages,
		"--channel-samples", *testWick.TestWick.ChannelSamples,
		"--hosted-zone", *testWick.TestWick.HostedZone,
		"--installation-affinity", *testWick.TestWick.AffinityType,
		"--installation-db-type", *testWick.TestWick.DBType,
		"--installation-filestore", *testWick.TestWick.FileStore,
		"--installation-size", *testWick.TestWick.Size,
		"--owner", *testWick.TestWick.Owner,
		"--provisioner", *testWick.TestWick.ProvisionerURL,
		"--samples", *testWick.TestWick.Samples,
	}
	cmd := exec.Command("/usr/local/bin/testwick", arguments...)
	impl.Log.Info("This is the command running", "command", cmd.String())

	var out bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &out
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		impl.Log.Error(err, "testwick installation creation failed")
		return v1alpha1.NotInjected, err
	}
	impl.Log.Info(stdout.String())
	return v1alpha1.Injected, nil
}

// Recover means the reconciler recovers the chaos action
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("Cleaning TestWick experiment")
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContianerRecordDecoder) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "testwickchaos",
		Object: &v1alpha1.TestWickChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("testwickchaos"),
			decoder: decoder,
		},
		ObjectList: &v1alpha1.TestWickChaosList{},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
