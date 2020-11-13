package main

import (
	"log"
	"os"

	"github.com/xanzy/go-gitlab"
)

func main() {
	token := os.Getenv("GITLAB_TOKEN")
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.Printf("My token : %s\n", token)
	client, err := gitlab.NewClient(token)
	if err != nil {
		panic(err)
	}

	projects, err := getProjects(client)
	if err != nil {
		panic(err)
	}
	for _, project := range projects {
		if project.Name != "marketplace" {
			continue
		}
		log.Printf("Project #%d %s [ %s ]\n", project.ID, project.Name, project.WebURL)
		// pipelines, err := getProjectPipelines(client, project)
		// if err != nil {
		// 	panic(err)
		// }

		// for _, pipeline := range pipelines {
		// 	log.Println(pipeline.String())
		// }
		jobs, err := getProjectJobs(client, project.ID)
		if err != nil {
			panic(err)
		}
		var durations []float64
		successCount := 0
		var lastStatus string
		for _, job := range jobs {
			// log.Printf("%s by %s at %s (%s) %fs\n", job.Commit.Message, job.Commit.AuthorName, job.Commit.CreatedAt, job.Status, job.Duration)
			durations = append(durations, job.Duration)
			if lastStatus == "" {
				lastStatus = job.Status
			}
			if job.Status != "success" {
				continue
			}
			successCount++
		}
		successRate := float64(successCount) / float64(len(jobs))
		averageDuration := func(durations []float64) float64 {
			totalDuration := 0.0
			for _, duration := range durations {
				totalDuration += duration
			}
			return totalDuration / float64(len(durations))
		}(durations)
		log.Printf("Success rate : %.2f%% [%d/%d]\n", successRate*100, successCount, len(jobs))
		log.Printf("Average job duration %.1fs\n", averageDuration)
		log.Printf("Last pipeline status : %s\n", lastStatus)
	}
}

func getProjects(client *gitlab.Client) ([]*gitlab.Project, error) {
	var projects []*gitlab.Project
	opts := &gitlab.ListProjectsOptions{
		Visibility: gitlab.Visibility(gitlab.PrivateVisibility),
	}
	projects, _, err := client.Projects.ListProjects(opts)

	return projects, err
}

func getProjectPipelines(client *gitlab.Client, project *gitlab.Project) ([]*gitlab.PipelineInfo, error) {
	opts := gitlab.ListProjectPipelinesOptions{}
	pipelines, _, err := client.Pipelines.ListProjectPipelines(project.ID, &opts)
	log.Printf("%+v\n", pipelines)

	return pipelines, err
}

func getProjectJobs(client *gitlab.Client, projectID int) ([]gitlab.Job, error) {
	service := client.Jobs
	opts := &gitlab.ListJobsOptions{}
	jobs, _, err := service.ListProjectJobs(projectID, opts)

	return jobs, err
}
