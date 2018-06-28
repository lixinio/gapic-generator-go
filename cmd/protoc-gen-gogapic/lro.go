// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

func isLRO(m *descriptor.MethodDescriptorProto) bool {
	// protoc puts a dot in front of name, signaling that the name is fully qualified.
	return *m.OutputType == ".google.longrunning.Operation"
}

func (g *generator) lroCall(servName string, m *descriptor.MethodDescriptorProto) {
	inType := g.types[*m.InputType]
	outType := g.types[*m.OutputType]
	inSpec := g.importSpec(inType)
	outSpec := g.importSpec(outType)

	lroType := lroTypeName(*m.Name)
	p := g.printf

	p("func (c *%sClient) %s(ctx context.Context, req *%s.%s, opts ...gax.CallOption) (*%s, error) {",
		servName, *m.Name, inSpec.name, *inType.Name, lroType)

	g.insertMetadata()
	g.appendCallOpts(m)
	p("  var resp *%s.%s", outSpec.name, *outType.Name)
	p("  err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {")
	p("    var err error")
	p("    resp, err = %s", grpcClientCall(servName, *m.Name))
	p("    return err")
	p("  }, opts...)")
	p("  if err != nil {")
	p("    return nil, err")
	p("  }")
	p("  return &%s{", lroType)
	p("    lro: longrunning.InternalNewOperation(c.LROClient, resp),")
	p("  }, nil")

	p("}")
	p("")

	g.imports[importSpec{path: "cloud.google.com/go/longrunning"}] = true
	g.imports[inSpec] = true
	g.imports[outSpec] = true
}

func (g *generator) lroType(servName string, m *descriptor.MethodDescriptorProto) {
	lroType := lroTypeName(*m.Name)
	p := g.printf

	// TODO(pongad): programmatically fill these.
	respType := "Foo.Bar"
	metaType := "Foo.MetaBar"
	hasMeta := true

	// Type definition
	{
		p("// %s manages a long-running operation from %s.", lroType, *m.Name)
		p("type %s struct {", lroType)
		p("  lro *longrunning.Operation")
		p("}")
		p("")
	}

	// LRO from name
	{
		p("// %[1]s returns a new %[1]s from a given name.", lroType)
		p("// The name must be that of a previously created %s, possibly from a different process.", lroType)
		p("func (c *%sClient) %[2]s(name string) *%[2]s {", servName, lroType)
		p("  return &%s{", lroType)
		p("    lro: longrunning.InternalNewOperation(c.LROClient, &longrunningpb.Operation{Name: name}),")
		p("  }")
		p("}")
		p("")

		g.imports[importSpec{name: "longrunningpb", path: "google.golang.org/genproto/googleapis/longrunning"}] = true
	}

	// Wait
	{
		p("// Wait blocks until the long-running operation is completed, returning the response and any errors encountered.")
		p("//")
		p("// See documentation of Poll for error-handling information.")
		p("func (op *%s) Wait(ctx context.Context, opts ...gax.CallOption) (*%s, error) {", lroType, respType)
		p("  var resp %s", respType)
		p("  if err := op.lro.WaitWithInterval(ctx, &resp, time.Minute, opts...); err != nil {")
		p("    return nil, err")
		p("  }")
		p("  return &resp, nil")
		p("}")
		p("")

		g.imports[importSpec{path: "time"}] = true
	}

	// Poll
	{
		p("// Poll fetches the latest state of the long-running operation.")
		p("//")
		if hasMeta {
			p("// Poll also fetches the latest metadata, which can be retrieved by Metadata.")
			p("//")
		}
		p("// If Poll fails, the error is returned and op is unmodified. If Poll succeeds and")
		p("// the operation has completed with failure, the error is returned and op.Done will return true.")
		p("// If Poll succeeds and the operation has completed successfully,")
		p("// op.Done will return true, and the response of the operation is returned.")
		p("// If Poll succeeds and the operation has not completed, the returned response and error are both nil.")
		p("func (op *%s) Poll(ctx context.Context, opts ...gax.CallOption) (*%s, error) {", lroType, respType)
		p("  var resp %s", respType)
		p("  if err := op.lro.Poll(ctx, &resp, opts...); err != nil {")
		p("    return nil, err")
		p("  }")
		p("  if !op.Done() {")
		p("    return nil, nil")
		p("  }")
		p("  return &resp, nil")
		p("}")
		p("")
	}

	// Metadata
	if hasMeta {
		p("// Metadata returns metadata associated with the long-running operation.")
		p("// Metadata itself does not contact the server, but Poll does.")
		p("// To get the latest metadata, call this method after a successful call to Poll.")
		p("// If the metadata is not available, the returned metadata and error are both nil.")
		p("func (op *%s) Metadata() (*%s, error) {", lroType, respType)
		p("  var meta %s", metaType)
		p("  if err := op.lro.Metadata(&meta); err == longrunning.ErrNoMetadata {")
		p("    return nil, nil")
		p("  } else if err != nil {")
		p("    return nil, err")
		p("  }")
		p("  return &meta, nil")
		p("}")
		p("")
	}

	// Done
	{
		p("// Done reports whether the long-running operation has completed.")
		p("func (op *%s) Done() bool {", lroType)
		p("return op.lro.Done()")
		p("}")
		p("")
	}

	// Name
	{
		p("// Name returns the name of the long-running operation.")
		p("// The name is assigned by the server and is unique within the service from which the operation is created.")
		p("func (op *%s) Name() string {", lroType)
		p("return op.lro.Name()")
		p("}")
		p("")
	}

}

func lroTypeName(methodName string) string {
	return methodName + "Operation"
}