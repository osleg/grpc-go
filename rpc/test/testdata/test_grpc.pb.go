/*
 *
 * Copyright 2014, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

package test

import (
	"fmt"
	"io"
	"github.com/grpc/grpc-go/rpc"
	context "golang.org/x/net/context"
	proto "github.com/golang/protobuf/proto"
)

type MathClient interface {
	Div(ctx context.Context, in *DivArgs, opts ...rpc.CallOption) (*DivReply, error)
	DivMany(ctx context.Context, opts ...rpc.CallOption) (Math_DivManyClient, error)
	Fib(ctx context.Context, m *FibArgs, opts ...rpc.CallOption) (Math_FibClient, error)
	Sum(ctx context.Context, opts ...rpc.CallOption) (Math_SumClient, error)
}

type mathClient struct {
	cc *rpc.ClientConn
}

func NewMathClient(cc *rpc.ClientConn) MathClient {
	return &mathClient{cc}
}

func (c *mathClient) Div(ctx context.Context, in *DivArgs, opts ...rpc.CallOption) (*DivReply, error) {
	out := new(DivReply)
	err := rpc.Invoke(ctx, "/test.Math/Div", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mathClient) DivMany(ctx context.Context, opts ...rpc.CallOption) (Math_DivManyClient, error) {
	stream, err := rpc.NewClientStream(ctx, c.cc, "/test.Math/DivMany", opts...)
	if err != nil {
		return nil, err
	}
	return &mathDivManyClient{stream}, nil
}

type Math_DivManyClient interface {
	Send(*DivArgs) error
	Recv() (*DivReply, error)
	rpc.ClientStream
}

type mathDivManyClient struct {
	rpc.ClientStream
}

func (x *mathDivManyClient) Send(m *DivArgs) error {
	return x.ClientStream.SendProto(m)
}

func (x *mathDivManyClient) Recv() (*DivReply, error) {
	m := new(DivReply)
	if err := x.ClientStream.RecvProto(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *mathClient) Fib(ctx context.Context, m *FibArgs, opts ...rpc.CallOption) (Math_FibClient, error) {
	stream, err := rpc.NewClientStream(ctx, c.cc, "/test.Math/Fib", opts...)
	if err != nil {
		return nil, err
	}
	x := &mathFibClient{stream}
	if err := x.ClientStream.SendProto(m); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Math_FibClient interface {
	Recv() (*Num, error)
	rpc.ClientStream
}

type mathFibClient struct {
	rpc.ClientStream
}

func (x *mathFibClient) Recv() (*Num, error) {
	m := new(Num)
	if err := x.ClientStream.RecvProto(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *mathClient) Sum(ctx context.Context, opts ...rpc.CallOption) (Math_SumClient, error) {
	stream, err := rpc.NewClientStream(ctx, c.cc, "/test.Math/Sum", opts...)
	if err != nil {
		return nil, err
	}
	return &mathSumClient{stream}, nil
}

type Math_SumClient interface {
	Send(*Num) error
	CloseAndRecv() (*Num, error)
	rpc.ClientStream
}

type mathSumClient struct {
	rpc.ClientStream
}

func (x *mathSumClient) Send(m *Num) error {
	return x.ClientStream.SendProto(m)
}

func (x *mathSumClient) CloseAndRecv() (*Num, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Num)
	if err := x.ClientStream.RecvProto(m); err != nil {
		return nil, err
	}
	// Read EOF.
	if err := x.ClientStream.RecvProto(m); err == io.EOF {
		return m, io.EOF
	}
	// gRPC protocol violation.
	return m, fmt.Errorf("Violate gRPC client streaming protocol: no EOF after the response.")
}


type MathServer interface {
	Div(context.Context, *DivArgs) (*DivReply, error)
	DivMany(Math_DivManyServer) error
	Fib(*FibArgs, Math_FibServer) error
	Sum(Math_SumServer) error
}

func RegisterService(s *rpc.Server, srv MathServer) {
	s.RegisterService(&_Math_serviceDesc, srv)
}

func _Math_Div_Handler(srv interface{}, ctx context.Context, buf []byte) (proto.Message, error) {
	in := new(DivArgs)
	if err := proto.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(MathServer).Div(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _Math_DivMany_Handler(srv interface{}, stream rpc.ServerStream) error {
	return srv.(MathServer).DivMany(&mathDivManyServer{stream})
}

type Math_DivManyServer interface {
	Send(*DivReply) error
	Recv() (*DivArgs, error)
	rpc.ServerStream
}

type mathDivManyServer struct {
	rpc.ServerStream
}

func (x *mathDivManyServer) Send(m *DivReply) error {
	return x.ServerStream.SendProto(m)
}

func (x *mathDivManyServer) Recv() (*DivArgs, error) {
	m := new(DivArgs)
	if err := x.ServerStream.RecvProto(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Math_Fib_Handler(srv interface{}, stream rpc.ServerStream) error {
	m := new(FibArgs)
	if err := stream.RecvProto(m); err != nil {
		return err
	}
	return srv.(MathServer).Fib(m, &mathFibServer{stream})
}

type Math_FibServer interface {
	Send(*Num) error
	rpc.ServerStream
}

type mathFibServer struct {
	rpc.ServerStream
}

func (x *mathFibServer) Send(m *Num) error {
	return x.ServerStream.SendProto(m)
}

func _Math_Sum_Handler(srv interface{}, stream rpc.ServerStream) error {
	return srv.(MathServer).Sum(&mathSumServer{stream})
}

type Math_SumServer interface {
	SendAndClose(*Num) error
	Recv() (*Num, error)
	rpc.ServerStream
}

type mathSumServer struct {
	rpc.ServerStream
}

func (x *mathSumServer) SendAndClose(m *Num) error {
	if err := x.ServerStream.SendProto(m); err != nil {
		return err
	}
	return nil
}

func (x *mathSumServer) Recv() (*Num, error) {
	m := new(Num)
	if err := x.ServerStream.RecvProto(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Math_serviceDesc = rpc.ServiceDesc{
	ServiceName: "test.Math",
	HandlerType: (*MathServer)(nil),
	Methods: []rpc.MethodDesc{
		{
			MethodName:	"Div",
			Handler:	_Math_Div_Handler,
		},
	},
	Streams: []rpc.StreamDesc{
		{
			StreamName:	"DivMany",
			Handler:	_Math_DivMany_Handler,
		},
		{
			StreamName:	"Fib",
			Handler:	_Math_Fib_Handler,
		},
		{
			StreamName:	"Sum",
			Handler:	_Math_Sum_Handler,
		},
	},
}


