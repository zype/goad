package request

import (
	"reflect"

	"github.com/gophergala2016/goad/cli/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/gophergala2016/goad/cli/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/awsutil"
)

//type Paginater interface {
//	HasNextPage() bool
//	NextPage() *Request
//	EachPage(fn func(data interface{}, isLastPage bool) (shouldContinue bool)) error
//}

// HasNextPage returns true if this request has more pages of data available.
func (r *Request) HasNextPage() bool {
	return len(r.nextPageTokens()) > 0
}

// nextPageTokens returns the tokens to use when asking for the next page of
// data.
func (r *Request) nextPageTokens() []interface{} {
	if r.Operation.Paginator == nil {
		return nil
	}

	if r.Operation.TruncationToken != "" {
		tr, _ := awsutil.ValuesAtPath(r.Data, r.Operation.TruncationToken)
		if len(tr) == 0 {
			return nil
		}

		switch v := tr[0].(type) {
		case *bool:
			if !aws.BoolValue(v) {
				return nil
			}
		case bool:
			if v == false {
				return nil
			}
		}
	}

	tokens := []interface{}{}
	tokenAdded := false
	for _, outToken := range r.Operation.OutputTokens {
		v, _ := awsutil.ValuesAtPath(r.Data, outToken)
		if len(v) > 0 {
			tokens = append(tokens, v[0])
			tokenAdded = true
		} else {
			tokens = append(tokens, nil)
		}
	}
	if !tokenAdded {
		return nil
	}

	return tokens
}

// NextPage returns a new Request that can be executed to return the next
// page of result data. Call .Send() on this request to execute it.
func (r *Request) NextPage() *Request {
	tokens := r.nextPageTokens()
	if len(tokens) == 0 {
		return nil
	}

	data := reflect.New(reflect.TypeOf(r.Data).Elem()).Interface()
	nr := New(r.Config, r.ClientInfo, r.Handlers, r.Retryer, r.Operation, awsutil.CopyOf(r.Params), data)
	for i, intok := range nr.Operation.InputTokens {
		awsutil.SetValueAtPath(nr.Params, intok, tokens[i])
	}
	return nr
}

// EachPage iterates over each page of a paginated request object. The fn
// parameter should be a function with the following sample signature:
//
//   func(page *T, lastPage bool) bool {
//       return true // return false to stop iterating
//   }
//
// Where "T" is the structure type matching the output structure of the given
// operation. For example, a request object generated by
// DynamoDB.ListTablesRequest() would expect to see dynamodb.ListTablesOutput
// as the structure "T". The lastPage value represents whether the page is
// the last page of data or not. The return value of this function should
// return true to keep iterating or false to stop.
func (r *Request) EachPage(fn func(data interface{}, isLastPage bool) (shouldContinue bool)) error {
	for page := r; page != nil; page = page.NextPage() {
		if err := page.Send(); err != nil {
			return err
		}
		if getNextPage := fn(page.Data, !page.HasNextPage()); !getNextPage {
			return page.Error
		}
	}

	return nil
}
