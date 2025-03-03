package v2ray_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	command "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"

	"github.com/anonopiran/Fly2Stats/internal/config"
	"github.com/anonopiran/Fly2Stats/internal/v2ray"
	mockCommand "github.com/anonopiran/Fly2Stats/mocks/github.com/xtls/xray-core/app/stats/command"
)

var _ = Describe("XrayServerType", func() {
	var (
		mockServiceClient *mockCommand.MockStatsServiceClient
		server            *v2ray.UpServer
	)

	BeforeEach(func() {
		mockServiceClient = mockCommand.NewMockStatsServiceClient(GinkgoT())
		server = v2ray.NewXrayServer(config.UpstreamUrlType{})
		server.IServer = &v2ray.XrayServerType{
			HandlerFactory: func(cci grpc.ClientConnInterface) command.StatsServiceClient {
				return mockServiceClient
			},
		}
	})

	Describe("QueryStats", func() {
		Context("when the connection is nil", func() {
			It("should return ErrNilConnection", func() {
				data, err := server.GetStats(context.Background(), nil)
				Expect(err).To(Equal(v2ray.ErrNilConnection))
				Expect(data).To(BeNil())
			})
		})

		Context("when the request is valid", func() {
			It("should return stats successfully", func() {
				mockServiceClient.EXPECT().QueryStats(mock.AnythingOfType("context.backgroundCtx"), mock.IsType(&command.QueryStatsRequest{})).Return(&command.QueryStatsResponse{}, nil)
				data, err := server.GetStats(context.Background(), &grpc.ClientConn{})
				Expect(err).To(BeNil())
				Expect(data).ToNot(BeNil())
			})
		})
	})

	Describe("ParseStats", func() {
		Context("when the stats are valid", func() {
			It("should parse stats correctly", func() {
				istat := [](v2ray.IStat){
					&command.Stat{Name: "user>>>testu1>>>traffic>>>downlink", Value: 10},
					&command.Stat{Name: "user>>>testu1>>>traffic>>>uplink", Value: 10},
					&command.Stat{Name: "user>>>testu2>>>traffic>>>downlink", Value: 10},
					&command.Stat{Name: "user>>>testu3>>>traffic>>>uplink", Value: 10},
				}
				result, err := server.IServer.ParseStats(istat)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeEmpty())
				_time := result[0].Time
				Expect(result).To(BeEquivalentTo([]v2ray.UserStatType{
					{Username: "testu1", Direction: v2ray.Downlink, Value: 10, Time: _time},
					{Username: "testu1", Direction: v2ray.Uplink, Value: 10, Time: _time},
					{Username: "testu2", Direction: v2ray.Downlink, Value: 10, Time: _time},
					{Username: "testu3", Direction: v2ray.Uplink, Value: 10, Time: _time},
				}))

			})
		})

		Context("when there are zero values", func() {
			It("should skip zero value stats", func() {
				istat := [](v2ray.IStat){
					&command.Stat{Name: "user>>>testu2>>>traffic>>>downlink", Value: 0},
					&command.Stat{Name: "user>>>testu3>>>traffic>>>uplink", Value: 0},
				}
				result, err := server.IServer.ParseStats(istat)
				Expect(err).To(BeNil())
				Expect(result).To(BeEmpty())
			})
		})

		Context("when regex fails to match", func() {
			It("should return without errors", func() {
				istat := []v2ray.IStat{
					&command.Stat{Name: "nomatch", Value: 0},
				}
				result, err := server.IServer.ParseStats(istat)
				Expect(err).To(BeNil())
				Expect(result).To(BeEmpty())
			})
		})
	})

	Describe("NewStatRequest", func() {
		Context("when called", func() {
			It("should return a valid request", func() {
				req, err := server.IServer.NewStatRequest()
				Expect(err).To(BeNil())
				Expect(req).ToNot(BeNil())
			})
		})
	})

	Describe("newServiceClient", func() {
		Context("when called with valid connection", func() {
			It("should return a StatsServiceClient", func() {
				client := server.IServer.NewServiceClient(&grpc.ClientConn{})
				Expect(client).ToNot(BeNil())
			})
		})
	})
})
