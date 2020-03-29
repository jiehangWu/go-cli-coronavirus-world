package main

import (
	"encoding/json"
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/urfave/cli"
	"github.com/graphql-go/graphql"
)

type country struct {
	date string `json:"date"`
	confirmed int `json:"confirmed`
	deaths int `json:"deaths"`
	recovered int `json:"recovered`
}

var data map[string]country

var countryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Country",
		Fields: graphql.Fields{
			"date": &graphql.Field{
				Type: graphql.String,
			},
			"confirmed": &graphql.Field{
				Type: graphql.Int,
			},
			"deaths": &graphql.Field{
				Type: graphql.Int,
			},
			"recovered": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"country": &graphql.Field{
				Type: countryType,
				Args: graphql.FieldConfigArgument{
					"confirmed": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					confirmedQuery, isOK := p.Args["id"].(string)
					if isOK {
						return data[confirmedQuery], nil
					}
					return nil, nil
				},
			},
		},
	},
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema: schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Println(result.Errors)
	}
	return result
}

func fetch(countryName string, option string, result interface{}) {
	url := "https://pomber.github.io/covid19/timeseries.json"
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Print("Error: ", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("Error: ", err)
		os.Exit(1)
	}

	err = json.Unmarshal(content, result)
	if err != nil {
		fmt.Print("Error: ", err)
	}
}

func main() {
	app := &cli.App{
		Name: "search",
		Usage: "search daily data for coronavirus",
		Action: func(c *cli.Context) error {
			country := ""
			if c.NArg() > 0 {
				country = c.Args().Get(0)
			}
			fetch(country, "", &data)
			http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request){
				result := executeQuery(r.URL.Query().Get("query"), schema)
				json.NewEncoder(w).Encode(result)
			})
			http.ListenAndServe(":3000", nil)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
