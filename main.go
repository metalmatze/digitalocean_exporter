package main

import (
	"context"
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func main() {
	tokenSource := &TokenSource{
		AccessToken: token,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	ctx := context.TODO()

	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{Page:1, PerPage:200})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("# DROPLETS")
	for _, droplet := range droplets {
		ip, _ := droplet.PublicIPv4()
		fmt.Printf("%s(%d:%d), %s, %s, %v\n", droplet.Name, droplet.Vcpus, droplet.Memory, ip, droplet.Created, droplet.Tags)
	}

	volumes, _, err := client.Storage.ListVolumes(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("# VOLUMES")

	var totalSize int64
	for _, vol := range volumes {
		totalSize = totalSize + vol.SizeGigaBytes
		fmt.Printf("%s(%d), %d, %v\n", vol.Name, vol.SizeGigaBytes, vol.DropletIDs, vol.CreatedAt)
	}

	fmt.Printf("You got a total of %d GB volumes for %0.2f€/month\n", totalSize, float64(totalSize)/10)
}
