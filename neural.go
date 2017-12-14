package deep

import (
	"fmt"

	"github.com/golang/glog"
)

type Neural struct {
	Layers []*Layer
	Biases [][]*Synapse
	Config *Config
	t      *Training
}

type Training struct {
	deltas    [][]float64
	oldDeltas [][]float64
}

type Config struct {
	Inputs     int
	Layout     []int
	Activation ActivationType
	Mode       Mode
	Weight     WeightInitializer `json:"-"`
	Error      ErrorMeasure      `json:"-"`
	Bias       float64
}

func NewNeural(c *Config) *Neural {

	if c.Weight == nil {
		c.Weight = NewUniform(0.5, 0)
	}
	if c.Activation == ActivationNone {
		c.Activation = ActivationSigmoid
	}

	layers := make([]*Layer, len(c.Layout))
	for i := range layers {
		act := c.Activation
		if i == (len(layers)-1) && c.Mode != ModeDefault {
			act = OutputActivation(c.Mode)
		}
		layers[i] = NewLayer(c.Layout[i], act)
	}

	for i := 0; i < len(layers)-1; i++ {
		layers[i].Connect(layers[i+1], c.Weight)
	}

	for _, neuron := range layers[0].Neurons {
		neuron.In = make([]*Synapse, c.Inputs)
		for i := range neuron.In {
			neuron.In[i] = NewSynapse(c.Weight())
		}
	}

	var biases [][]*Synapse
	if c.Bias > 0 {
		biases = make([][]*Synapse, len(layers))
		for i := 0; i < len(layers); i++ {
			if c.Mode == ModeRegression && i == len(layers)-1 {
				continue
			}
			biases[i] = layers[i].ApplyBias(c.Weight)
		}
	}

	return &Neural{
		Layers: layers,
		Biases: biases,
		Config: c,
		t:      getTraining(layers),
	}
}

func getTraining(layers []*Layer) *Training {
	deltas := make([][]float64, len(layers))
	oldDeltas := make([][]float64, len(layers))
	for i, l := range layers {
		deltas[i] = make([]float64, len(l.Neurons))
		oldDeltas[i] = make([]float64, len(l.Neurons)*len(l.Neurons[0].In))
	}
	return &Training{
		deltas:    deltas,
		oldDeltas: oldDeltas,
	}
}

func (n *Neural) Fire() {
	for i := range n.Biases {
		for j := range n.Biases[i] {
			n.Biases[i][j].Fire(n.Config.Bias)
		}
	}
	for _, l := range n.Layers {
		l.Fire()
	}
}

func (n *Neural) set(input []float64) {
	if len(input) != n.Config.Inputs {
		glog.Errorf("Invalid input dimension - expected: %d got: %d", n.Config.Inputs, len(input))
	}
	for _, n := range n.Layers[0].Neurons {
		for i := 0; i < len(input); i++ {
			n.In[i].Fire(input[i])
		}
	}
}

func (n *Neural) Forward(input []float64) {
	n.set(input)
	n.Fire()
}

func (n *Neural) Predict(input []float64) []float64 {
	n.Forward(input)

	outLayer := n.Layers[len(n.Layers)-1]
	out := make([]float64, len(outLayer.Neurons))
	for i, neuron := range outLayer.Neurons {
		out[i] = neuron.Value
	}
	return out
}

func (n *Neural) String() string {
	var s string
	for _, l := range n.Layers {
		s = fmt.Sprintf("%s\n%s", s, l)
	}
	return s
}
