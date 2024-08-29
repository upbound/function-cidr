package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/function-sdk-go/resource"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1beta1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1beta1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"cidr-host": {
			reason: "should return the CIDR host of the request",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrhost",
								},
							},
							"prefix": {
								Kind: &structpb.Value_StringValue{
									StringValue: "127.0.0.0/24",
								},
							},
							"hostNum": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 111,
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": {"cidr": "127.0.0.111"}}}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"cidr-subnet": {
			reason: "should return the cidr subnet of the request",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrsubnet",
								},
							},
							"prefix": {
								Kind: &structpb.Value_StringValue{
									StringValue: "127.0.0.0/24",
								},
							},
							"newBits": {
								Kind: &structpb.Value_ListValue{
									ListValue: &structpb.ListValue{
										Values: []*structpb.Value{
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8,
												},
											},
										},
									},
								},
							},
							"netNum": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 3,
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": {"cidr": "127.0.0.3/32"}}}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"cidr-netmask": {
			reason: "should return the CIDR netmask of the request",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrnetmask",
								},
							},
							"prefix": {
								Kind: &structpb.Value_StringValue{
									StringValue: "127.0.0.0/24",
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": {"cidr": "255.255.255.0"}}}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"cidr-subnets": {
			reason: "should return the cidr subnet of the request",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrsubnets",
								},
							},
							"prefix": {
								Kind: &structpb.Value_StringValue{
									StringValue: "127.0.0.0/24",
								},
							},
							"newBits": {
								Kind: &structpb.Value_ListValue{
									ListValue: &structpb.ListValue{
										Values: []*structpb.Value{
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8,
												},
											},
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4,
												},
											},
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 2,
												},
											},
										},
									},
								},
							},
							"netNum": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 3,
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": {"cidr": ["127.0.0.0/32", "127.0.0.16/28", "127.0.0.64/26"]}}}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"cidr-subnetloop": {
			reason: "should return the cidr subnet of the request",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrsubnetloop",
								},
							},
							"prefix": {
								Kind: &structpb.Value_StringValue{
									StringValue: "10.0.0.0/24",
								},
							},
							"newBits": {
								Kind: &structpb.Value_ListValue{
									ListValue: &structpb.ListValue{
										Values: []*structpb.Value{
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8,
												},
											},
										},
									},
								},
							},
							"netNumCount": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 3,
								},
							},
							"offset": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 48,
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": {"cidr": ["10.0.0.48/32", "10.0.0.49/32", "10.0.0.50/32"]}}}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"multi-prefix-loop": {
			reason: "should return multiple cidr subnets for the request",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "multiprefixloop",
								},
							},
							"multiPrefix": {
								Kind: &structpb.Value_ListValue{
									ListValue: &structpb.ListValue{
										Values: []*structpb.Value{
											{
												Kind: &structpb.Value_StructValue{
													StructValue: &structpb.Struct{
														Fields: map[string]*structpb.Value{
															"prefix": {
																Kind: &structpb.Value_StringValue{
																	StringValue: "10.10.0.0/24",
																},
															},
															"newBits": {
																Kind: &structpb.Value_ListValue{
																	ListValue: &structpb.ListValue{
																		Values: []*structpb.Value{
																			{
																				Kind: &structpb.Value_NumberValue{
																					NumberValue: 8,
																				},
																			},
																			{
																				Kind: &structpb.Value_NumberValue{
																					NumberValue: 4,
																				},
																			},
																			{
																				Kind: &structpb.Value_NumberValue{
																					NumberValue: 2,
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Kind: &structpb.Value_StructValue{
													StructValue: &structpb.Struct{
														Fields: map[string]*structpb.Value{
															"prefix": {
																Kind: &structpb.Value_StringValue{
																	StringValue: "10.12.0.0/24",
																},
															},
															"newBits": {
																Kind: &structpb.Value_ListValue{
																	ListValue: &structpb.ListValue{
																		Values: []*structpb.Value{
																			{
																				Kind: &structpb.Value_NumberValue{
																					NumberValue: 4,
																				},
																			},
																			{
																				Kind: &structpb.Value_NumberValue{
																					NumberValue: 4,
																				},
																			},
																			{
																				Kind: &structpb.Value_NumberValue{
																					NumberValue: 4,
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": ` +
								`{"cidr": {"10.10.0.0/24": ["10.10.0.0/32", "10.10.0.16/28", "10.10.0.64/26"],` +
								`"10.12.0.0/24": ["10.12.0.0/28", "10.12.0.16/28", "10.12.0.32/28"]}}}}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"cidr-subnets-from-context": {
			reason: "should return the cidr subnet of the prefixField retrieved from the context",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Context: &structpb.Struct {
						Fields: map[string]*structpb.Value {
							"apiextensions.crossplane.io/extra-resources": structpb.NewStructValue(resource.MustStructJSON(`{
									"XCluster": [
        							    {
        							        "apiVersion": "example.crossplane.io/v1",
        							        "kind": "XCluster",
        							        "spec": {
        							            "cidrBlock": "10.0.0.0/20"
        							        }
        							    }
        							]

							}`)),
						},
					},
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrsubnets",
								},
							},
							"prefixField": {
								Kind: &structpb.Value_StringValue{
									StringValue: "context.apiextensions\\.crossplane\\.io/extra-resources.XCluster.0.spec.cidrBlock",
								},
							},
							"newBits": {
								Kind: &structpb.Value_ListValue{
									ListValue: &structpb.ListValue{
										Values: []*structpb.Value{
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 1,
												},
											},
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 1,
												},
											},
										},
									},
								},
							},
							"outputField": {
								Kind: &structpb.Value_StringValue{
									StringValue: "status.atFunction.cidr.partitions",
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{"apiVersion":"","kind":"","status": {"atFunction": {"cidr": {"partitions": ["10.0.0.0/21", "10.0.8.0/21"]}}}}`),
						},
					},
					Context: &structpb.Struct {
						Fields: map[string]*structpb.Value {
							"apiextensions.crossplane.io/extra-resources": structpb.NewStructValue(resource.MustStructJSON(`{
									"XCluster": [
        							    {
        							        "apiVersion": "example.crossplane.io/v1",
        							        "kind": "XCluster",
        							        "spec": {
        							            "cidrBlock": "10.0.0.0/20"
        							        }
        							    }
        							]

							}`)),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

		"cidr-subnets-from-desired": {
			reason: "should return the cidr subnet of the prefixField retrieved from the desired field",
			args: args{
				ctx: context.Background(),
				req: &fnv1beta1.RunFunctionRequest{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "platform.upbound.io/v1alpha1",
								"kind": "XCIDR",
                                "status": {
									"atFunction": {
										"cidr": {
											"partitions": ["10.0.0.0/21"]
										}
									}
								}
							}`),
						},
					},
					Input: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"cidrFunc": {
								Kind: &structpb.Value_StringValue{
									StringValue: "cidrsubnets",
								},
							},
							"prefixField": {
								Kind: &structpb.Value_StringValue{
									StringValue: "desired.composite.resource.status.atFunction.cidr.partitions[0]",
								},
							},
							"newBits": {
								Kind: &structpb.Value_ListValue{
									ListValue: &structpb.ListValue{
										Values: []*structpb.Value{
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 1,
												},
											},
											{
												Kind: &structpb.Value_NumberValue{
													NumberValue: 1,
												},
											},
										},
									},
								},
							},
							"outputField": {
								Kind: &structpb.Value_StringValue{
									StringValue: "status.atFunction.cidr.private.subnets",
								},
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion":"",
								"kind":"",
								"status": {
									"atFunction": {
										"cidr": {
											"private": {
												"subnets": ["10.0.0.0/22", "10.0.4.0/22"]
											},
											"partitions": ["10.0.0.0/21"]
										}
									}
								}
							}`),
						},
					},
					Meta: &fnv1beta1.ResponseMeta{
						Ttl: &durationpb.Duration{
							Seconds: 60,
						},
					},
				},
				err: nil,
			},
		},

	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
