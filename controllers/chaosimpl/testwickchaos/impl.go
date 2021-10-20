package testwickchaos

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"sync"

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
		"--provisioner-image-name", *testWick.TestWick.ProvisionerImageName,
		"--provisioner-image-tag", *testWick.TestWick.ProvisionerImageTag,
		"--channel-messages-sleep", *testWick.TestWick.ChannelMessagesSleep,
	}
	cmd := exec.Command("/usr/local/bin/testwick", arguments...)
	impl.Log.Info("This is the command running", "command", cmd.String())

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rStdout, wStdout := io.Pipe()
	rStderr, wStderr := io.Pipe()

	cmd.Stdout = wStdout
	cmd.Stderr = wStderr

	var wg sync.WaitGroup

	// Log and buffer stdout.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bufferAndLog(rStdout, stdout, impl.Log); err != nil {
			impl.Log.Error(err, "failed to scan stdout")
		}
	}()

	// Log and buffer stderr.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bufferAndLog(rStderr, stderr, impl.Log); err != nil {
			impl.Log.Error(err, "failed to scan stderr")
		}
	}()

	var err error
	go func() {
		err = cmd.Run()
		wStdout.Close()
		wStderr.Close()
	}()

	wg.Wait()

	if err != nil {
		impl.Log.Error(err, "failed invocation")

		return v1alpha1.NotInjected, err
	}

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

func bufferAndLog(reader io.Reader, buffer *bytes.Buffer, logger logr.Logger) error {
	scanner := bufio.NewScanner(io.TeeReader(reader, buffer))
	for scanner.Scan() {
		logger.Info(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
