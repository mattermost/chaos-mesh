package testwickchaos

import (
	"context"

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
