package linear

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
    "math"

	"github.com/cdipaolo/goml/base"
)

// Logistic represents the logistic classification
// model with a sigmoidal hypothesis
//
// https://en.wikipedia.org/wiki/Logistic_regression
//
// The model is currently optimized using Gradient
// Ascent, not Newton's method, etc.
type Logistic struct {
	// alpha and maxIterations are used only for
	// GradientAscent during learning. If maxIterations
	// is 0, then GradientAscent will run until the
	// algorithm detects convergance.
	//
	// regularization is used as the regularization
	// term to avoid overfitting within regression.
	// Having a regularization term of 0 is like having
	// _no_ data regularization. The higher the term,
	// the greater the bias on the regression
	alpha          float64
	regularization float64
	maxIterations  int

	// trainingSet and expectedResults are the
	// 'x', and 'y' of the data, expressed as
	// vectors, that the model can optimize from
	trainingSet     [][]float64
	expectedResults []float64

	Parameters []float64 `json:"theta"`
}

// NewLogistic takes in a learning rate alpha, a regularization
// parameter value (0 means no regularization, higher value
// means higher bias on the model,) the maximum number of
// iterations the data can go through in gradient descent,
// as well as a training set and expected results for that
// training set.
func NewLogistic(alpha, regularization float64, maxIterations int, trainingSet [][]float64, expectedResults []float64) *Logistic {
	var params []float64
	if trainingSet == nil || len(trainingSet) == 0 {
		params = []float64{}
	} else {
		params = make([]float64, len((trainingSet)[0])+1)
	}

	return &Logistic{
		alpha:          alpha,
		regularization: regularization,
		maxIterations:  maxIterations,

		trainingSet:     trainingSet,
		expectedResults: expectedResults,

		// initialize θ as the zero vector (that is,
		// the vector of all zeros)
		Parameters: params,
	}
}

// UpdateTrainingSet takes in a new training set (variable x)
// as well as a new result set (y). This could be useful if
// you want to retrain a model starting with the parameter
// vector of a previous training session, but most of the time
// wouldn't be used.
func (l *Logistic) UpdateTrainingSet(trainingSet [][]float64, expectedResults []float64) error {
	if len(trainingSet) == 0 {
		return fmt.Errorf("Error: length of given training set is 0! Need data!")
	}
	if len(expectedResults) == 0 {
		return fmt.Errorf("Error: length of given result data set is 0! Need expected results!")
	}

	l.trainingSet = trainingSet
	l.expectedResults = expectedResults

	return nil
}

// UpdateLearningRate set's the learning rate of the model
// to the given float64.
func (l *Logistic) UpdateLearningRate(a float64) {
	l.alpha = a
}

// LearningRate returns the learning rate α for gradient
// descent to optimize the model. Could vary as a function
// of something else later, potentially.
func (l *Logistic) LearningRate() float64 {
	return l.alpha
}

// MaxIterations returns the number of maximum iterations
// the model will go through in GradientAscent, in the
// worst case
func (l *Logistic) MaxIterations() int {
	return l.maxIterations
}

// Predict takes in a variable x (an array of floats,) and
// finds the value of the hypothesis function given the
// current parameter vector θ
func (l *Logistic) Predict(x []float64) ([]float64, error) {
	if len(x)+1 != len(l.Parameters) {
		return nil, fmt.Errorf("Error: Parameter vector should be 1 longer than input vector!\n\tLength of x given: %v\n\tLength of parameters: %v\n", len(x), len(l.Parameters))
	}

	// include constant term in sum
	sum := l.Parameters[0]

	for i := range x {
		sum += x[i] * l.Parameters[i+1]
	}

	result := 1 / (1 + math.Exp(-sum))

	return []float64{result}, nil
}

// Learn takes the struct's dataset and expected results and runs
// batch gradient descent on them, optimizing theta so you can
// predict based on those results
func (l *Logistic) Learn() error {
	if l.trainingSet == nil || l.expectedResults == nil {
		err := fmt.Errorf("ERROR: Attempting to learn with no training examples!\n")
		fmt.Printf(err.Error())
		return err
	}

	examples := len(l.trainingSet)
	if examples == 0 || len(l.trainingSet[0]) == 0 {
		err := fmt.Errorf("ERROR: Attempting to learn with no training examples!\n")
		fmt.Printf(err.Error())
		return err
	}
	if len(l.expectedResults) == 0 {
		err := fmt.Errorf("ERROR: Attempting to learn with no expected results! This isn't an unsupervised model!! You'll need to include data before you learn :)\n")
		fmt.Printf(err.Error())
		return err
	}

	fmt.Printf("Training:\n\tModel: Logistic (Binary) Classification\n\tOptimization Method: Batch Gradient Descent\n\tTraining Examples: %v\n\tFeatures: %v\n\tLearning Rate α: %v\n\tRegularization Parameter λ: %v\n...\n\n", examples, len(l.trainingSet[0]), l.alpha, l.regularization)

	err := base.GradientAscent(l)
	if err != nil {
		fmt.Printf("\nERROR: Error while learning –\n\t%v\n\n", err)
		return err
	}

	fmt.Printf("Training Completed.\n%v\n\n", l)
	return nil
}

// String implements the fmt interface for clean printing. Here
// we're using it to print the model as the equation h(θ)=...
// where h is the logistic hypothesis model
func (l *Logistic) String() string {
	features := len(l.Parameters) - 1
	if len(l.Parameters) == 0 {
		fmt.Printf("ERROR: Attempting to print model with the 0 vector as it's parameter vector! Train first!\n")
	}
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("h(θ,x) = 1 / (1 + exp(-θx))\nθx = %.3f + ", l.Parameters[0]))

	length := features + 1
	for i := 1; i < length; i++ {
		buffer.WriteString(fmt.Sprintf("%.5f(x[%d])", l.Parameters[i], i))

		if i != features {
			buffer.WriteString(fmt.Sprintf(" + "))
		}
	}

	return buffer.String()
}

// Dj returns the partial derivative of the cost function J(θ)
// with respect to theta[j] where theta is the parameter vector
// associated with our hypothesis function Predict (upon which
// we are optimizing
func (l *Logistic) Dj(j int) (float64, error) {
	if j > len(l.Parameters)-1 {
		return 0, fmt.Errorf("J (%v) would index out of the bounds of the training set data (len: %v)", j, len(l.Parameters))
	}

	var sum float64

	for i := range l.trainingSet {
		prediction, err := l.Predict(l.trainingSet[i])
		if err != nil {
			return 0, err
		}

		// account for constant term
		// x is x[i][j] via Andrew Ng's terminology
		var x float64
		if j == 0 {
			x = 1
		} else {
			x = l.trainingSet[i][j-1]
		}

		sum += (l.expectedResults[i] - prediction[0]) * x
	}

	// add in the regularization term
	// λ*θ[j]
	//
	// notice that we don't count the
	// constant term
	if j != 0 {
		sum += l.regularization * l.Parameters[j]
	}

	return sum, nil
}

// Theta returns the parameter vector θ for use in persisting
// the model, and optimizing the model through gradient descent
// ( or other methods like Newton's Method)
func (l *Logistic) Theta() []float64 {
	return l.Parameters
}

// PersistToFile takes in an absolute filepath and saves the
// parameter vector θ to the file, which can be restored later.
// The function will take paths from the current directory, but
// functions
//
// The data is stored as JSON because it's one of the most
// efficient storage method (you only need one comma extra
// per feature + two brackets, total!) And it's extendable.
func (l *Logistic) PersistToFile(path string) error {
	if path == "" {
		return fmt.Errorf("ERROR: you just tried to persist your model to a file with no path!! That's a no-no. Try it with a valid filepath")
	}

	bytes, err := json.Marshal(l.Parameters)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// RestoreFromFile takes in a path to a parameter vector theta
// and assigns the model it's operating on's parameter vector
// to that.
//
// The path must ba an absolute path or a path from the current
// directory
//
// This would be useful in persisting data between running
// a model on data, or for graphing a dataset with a fit in
// another framework like Julia/Gadfly.
func (l *Logistic) RestoreFromFile(path string) error {
	if path == "" {
		return fmt.Errorf("ERROR: you just tried to restore your model from a file with no path! That's a no-no. Try it with a valid filepath")
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &l.Parameters)
	if err != nil {
		return err
	}

	return nil
}
