package testwickchaos

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	tw "github.com/mattermost/mattermost-cloud/cmd/tools/testwick"
	cmodel "github.com/mattermost/mattermost-cloud/model"
	mmodel "github.com/mattermost/mattermost-server/v5/model"
	"github.com/sirupsen/logrus"

	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
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
	testWickConfig := obj.(*v1alpha1.TestWickChaos)
	impl.Log.Info("Creating installations", "installations", testWickConfig.TestWick.Samples)

	u, err := url.Parse(*testWickConfig.TestWick.ProvisionerURL)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	provisionerClient := cmodel.NewClient(u.String())
	g, _ := errgroup.WithContext(context.Background())

	for i := 0; i < *testWickConfig.TestWick.Samples; i++ {

		// dnsName := fmt.Sprintf("%s.%s", tw.normalizeDNS(namesgenerator.GetRandomName(5)), *testWickConfig.TestWick.HostedZone)

		g.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			mmClient := mmodel.NewAPIv4Client(fmt.Sprintf("https://%s", *testWickConfig.TestWick.HostedZone))
			testwicker := tw.NewTestWicker(provisionerClient, mmClient, logrus.Logger)

			workflow := tw.NewWorkflow(logrus.Logger)
			workflow.AddStep(Step{
				Name: "CreateInstallation",
				Func: testwicker.CreateInstallation(&cmodel.CreateInstallationRequest{
					DNS:       *testWickConfig.TestWick.HostedZone,
					OwnerID:   *testWickConfig.TestWick.Owner,
					Size:      *testWickConfig.TestWick.Size,
					Affinity:  *testWickConfig.TestWick.AffinityType,
					Database:  *testWickConfig.TestWick.DBType,
					Filestore: *testWickConfig.TestWick.FileStore,
				}),
			}, Step{
				Name: "WaitForInstallationStable",
				Func: testwicker.WaitForInstallationStable(),
			}, Step{
				Name: "IsInstallationUp",
				Func: testwicker.IsUP(),
			}, Step{
				Name: "SetupInstallation",
				Func: testwicker.SetupInstallation(),
			}, Step{
				Name: "CreateTeam",
				Func: testwicker.CreateTeam(),
			}, Step{
				Name: "AddTeamMember",
				Func: testwicker.AddTeamMember(),
			})

			for j := 0; j < *testWickConfig.TestWick.ChannelSamples; j++ {
				workflow.AddStep(Step{
					Name: "CreateChannel",
					Func: testwicker.CreateChannel(),
				}, Step{
					Name: "CreateIncomingWebhook",
					Func: testwicker.CreateIncomingWebhook(),
				}, Step{
					Name: "PostMessages",
					Func: testwicker.PostMessage(*testWickConfig.TestWick.ChannelMessages),
				})
			}
			workflow.AddStep(Step{
				Name: "DeleteInstallation",
				Func: testwicker.DeleteInstallation(),
			})
			if err := workflow.Run(testwicker, ctx); err != nil {
				testwicker.DeleteInstallation()
				cancel()
				return err
			}

			return nil
		})
	}

	return v1alpha1.Injected, nil
}

// Recover means the reconciler recovers the chaos action
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("Cleaning TestWick experiment")
	testWick := obj.(*v1alpha1.TestWickChaos)
	impl.Log.Info("Cleaning installations", "installations", testWick.TestWick.Samples)
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
