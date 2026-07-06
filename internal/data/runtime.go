package data

import (
        "errors"
        "fmt"
        "strconv"
        "strings"
)

var ErrInvalidRuntimeForat = errors.New("Invalid runtime format")

type Runtime int32

func(r Runtime) MarshalJSON() ([]byte, error){
  jsonValue := fmt.Sprintf("%d mins", r)

  Quotedjsonvalue := strconv.Quote(jsonValue)

  return []byte(Quotedjsonvalue), nil
}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error{

  unquotejsonval, err := strconv.Unquote(string(jsonValue))
  if err != nil{
    return ErrInvalidRuntimeForat
  }
  parts := strings.Split(unquotejsonval, " ")

  if len(parts) != 2 || parts[1] != "mins"{
    return ErrInvalidRuntimeForat
  } 

  i, err := strconv.ParseInt(parts[0], 10, 32)
  if err != nil{
    return ErrInvalidRuntimeForat
  }

  *r = Runtime(i)

  return nil
}