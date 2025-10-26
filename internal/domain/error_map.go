
package domain

type Failure struct {
    Status    Status   
    ErrorCode ErrorCode 
}

type ErrorClassifier interface {
    Classify(err error, ctxErr error) Failure
}
