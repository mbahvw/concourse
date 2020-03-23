package integration_test

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/concourse/concourse/atc"
	concourse "github.com/concourse/concourse/go-concourse/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
)

var basicPipelineConfig = []byte(`
---
jobs:
- name: simple
`)

var _ = Describe("ATC Integration Test", func() {
	var (
		atcProcess ifrit.Process
		atcURL     string
		client     concourse.Client
	)

	BeforeEach(func() {
		atcURL = fmt.Sprintf("http://localhost:%v", cmd.BindPort)

		runner, err := cmd.Runner([]string{})
		Expect(err).NotTo(HaveOccurred())

		atcProcess = ifrit.Invoke(runner)

		Eventually(func() error {
			_, err := http.Get(atcURL + "/api/v1/info")
			return err
		}, 20*time.Second).ShouldNot(HaveOccurred())

		client = login(atcURL, "test", "test")
		givenAPipeline(client, "pipeline")
		whenIArchiveIt(client, "pipeline")
	})

	AfterEach(func() {
		atcProcess.Signal(os.Interrupt)
		<-atcProcess.Wait()
	})

	It("can archive pipelines", func() {
		pipeline := getPipeline(client, "pipeline")
		Expect(pipeline.Archived).To(BeTrue(), "pipeline was not archived")
		Expect(pipeline.Paused).To(BeTrue(), "pipeline was not paused")
	})

	It("fails when unpausing an archived pipeline", func() {
		_, err := client.Team("main").UnpausePipeline("pipeline")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("archived pipelines cannot be unpaused"))
	})

	It("archived pipelines can have their name re-used", func() {
		_, version, _, _ := client.Team("main").PipelineConfig("pipeline")
		client.Team("main").CreateOrUpdatePipelineConfig("pipeline", version, basicPipelineConfig, false)
		pipeline := getPipeline(client, "pipeline")
		Expect(pipeline.Archived).To(BeFalse(), "pipeline is still archived")
		Expect(pipeline.Paused).To(BeTrue(), "pipeline was not paused")
	})
})

func givenAPipeline(client concourse.Client, pipelineName string) {
	_, _, _, err := client.Team("main").CreateOrUpdatePipelineConfig(pipelineName, "0", basicPipelineConfig, false)
	Expect(err).NotTo(HaveOccurred())
}

func whenIArchiveIt(client concourse.Client, pipelineName string) {
	httpClient := client.HTTPClient()
	request, _ := http.NewRequest(
		"PUT",
		client.URL()+"/api/v1/teams/main/pipelines/"+pipelineName+"/archive",
		nil,
	)
	_, err := httpClient.Do(request)
	Expect(err).ToNot(HaveOccurred())
}

func getPipeline(client concourse.Client, pipelineName string) atc.Pipeline {
	pipeline, _, err := client.Team("main").Pipeline(pipelineName)
	Expect(err).ToNot(HaveOccurred())
	return pipeline
}
