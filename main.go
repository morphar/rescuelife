package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/cheggaaa/pb"
	"github.com/morphar/rescuelife/minicrawler"
)

type Media struct {
	Id             string `json:"id"`
	MediaType      string `json:"media_type"`
	Format         string `json:"format"`
	Processed      bool   `json:"processed"`
	CreatedAt      int    `json:"created_at"`
	UpdatedAt      int    `json:"updated_at"`
	TakenAt        int    `json:"taken_at"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Orientation    int    `json:"orientation"`
	Privacy        int    `json:"privacy"`
	IsBestPhoto    bool   `json:"is_best_photo"`
	TimeZoneOffset int    `json:"time_zone_offset"`
	Hidden         bool   `json:"hidden"`
	Visible        bool   `json:"visible"`
	Filesize       int    `json:"filesize"`
	BucketId       int    `json:"bucket_id"`
	Status         string `json:"status"`
	Retries        int    `json:"retries"`
}

type APIResponse struct {
	Status       int     `json:"status"`
	Media        []Media `json:"media"`
	Total        int     `json:"total"`
	Limit        int     `json:"limit"`
	Offset       int     `json:"offset"`
	UsingCache   bool    `json:"using_cache"`
	ResponseTime int     `json:"response_time"`
}

var (
	loginUrl    *url.URL
	signinUrl   *url.URL
	apiPageUrl  *url.URL
	apiUrl      *url.URL
	originalUrl *url.URL

	signinValues url.Values

	accessTokenRE *regexp.Regexp
	accessToken   string

	pathPerm os.FileMode = 0770
	filePerm os.FileMode = 0770

	mediaPath string = "picturelife"
	indexPath string = "pl_index.json"

	// Flags
	retryFlag  bool = false // Retry failed images and videos?
	helpFlag   bool = false // Retry failed images and videos?
	statusFlag bool = false // Retry failed images and videos?
)

func init() {
	var err error

	flag.BoolVar(&retryFlag, "retry", retryFlag, "Retry failed images and videos?")
	flag.BoolVar(&helpFlag, "help", helpFlag, "Print help text")
	flag.BoolVar(&statusFlag, "status", statusFlag, "Print out current status")

	loginUrl, err = url.Parse("http://picturelife.com/login")
	if err != nil {
		panic("Unable to parse login URL")
	}

	// Login posts to this
	signinUrl, err = url.Parse("http://picturelife.com/signin")
	if err != nil {
		panic("Unable to parse sign in URL")
	}

	apiPageUrl, err = url.Parse("http://picturelife.com/api")
	if err != nil {
		panic("Unable to parse API Page URL")
	}

	originalUrl, err = url.Parse("http://picturelife.com/d/original/")
	if err != nil {
		panic("Unable to parse API Page URL")
	}

	accessTokenRE = regexp.MustCompile("<script>\\s*pl\\.access_token\\s*=\\s*'([^']+)';\\s*pl\\.api_url\\s*=\\s*'([^']+)'\\s*</script>")

	err = os.MkdirAll(mediaPath, pathPerm)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	if helpFlag {
		printHelp()
		return
	}

	if statusFlag {
		printStatus()
		return
	}

	// Instantiate the crawler
	client := minicrawler.New()

	// Ask for email and password
	signinValues := getCredentials()

	res := client.GetOrDie(loginUrl.String())
	res.Body.Close()

	res = client.PostFormOrDie(signinUrl.String(), signinValues)
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if strings.Contains(string(body), "Login error! Please check your email and password.") {
		fmt.Println("Login error! Please check your email and password.")
		return
	}

	res = client.GetOrDie(apiPageUrl.String())
	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	fmt.Print("Trying to extract Access Token and API URL...")
	parts := accessTokenRE.FindStringSubmatch(string(body))
	if len(parts) != 3 {
		fmt.Println("\nUnable to extract Access Token and API URL.")
		fmt.Println("This is the source code received:")
		fmt.Println(string(body))
		return
	}
	fmt.Println(" Done!")

	accessToken = parts[1]
	apiUrl, err = url.Parse(parts[2])
	if err != nil {
		fmt.Println("Unable to parse API Page URL")
		return
	}

	// So far, so good... Now extract the index json, if it hasn't already been done

	// If the JSON index file does not exist, we'll fetch it from the API and create it

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("\nTrying to extract index of all files...")

		var progress *pb.ProgressBar
		var allMedia []Media

		indexUrl := apiUrl.String() + "/media/index"

		offset := 0
		limit := 500
		total := -1

		formValues := url.Values{
			"taken_at_after":      {"0"},
			"include_hidden":      {"true"},
			"show_invisible":      {"true"},
			"warm_thumbs":         {"false"},
			"include_names":       {"false"},
			"include_comments":    {"false"},
			"include_signature":   {"false"},
			"include_access_info": {"false"},
			"include_likes":       {"false"},
			"offset":              {strconv.Itoa(offset)},
			"limit":               {strconv.Itoa(limit)},
			"access_token":        {accessToken},
		}

		for total == -1 || offset < total {
			formValues.Set("offset", strconv.Itoa(offset))

			res := client.PostFormOrDie(indexUrl, formValues)
			body, err = ioutil.ReadAll(res.Body)
			res.Body.Close()

			var apiResponse APIResponse
			err := json.Unmarshal(body, &apiResponse)
			if err != nil {
				fmt.Println("ERROR! Unable to read JSON response from API. Please try again later.")
				os.Exit(0)
			}

			allMedia = append(allMedia, apiResponse.Media...)
			total = apiResponse.Total

			if progress == nil {
				progress = pb.New(total)
				progress.ShowCounters = true
				progress.ShowTimeLeft = true
				progress.Start()
			}

			progress.Set(offset)

			offset += limit
		}

		progress.FinishPrint("Done fetching JSON index")

		mediaJson, _ := json.Marshal(allMedia)
		err = ioutil.WriteFile(indexPath, mediaJson, filePerm)

		if err != nil {
			fmt.Println("ERROR! Unable to write JSON index file to disk. Sorry...")
			fmt.Println("Please go to GitHub and open an issue.")
			os.Exit(0)
		}
	}

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("ERROR! Unable to find the JSON index file from disk. Sorry...")
		fmt.Println("Please go to GitHub and open an issue.")
		os.Exit(0)
	}

	src, err := ioutil.ReadFile(indexPath)
	if err != nil {
		fmt.Println("ERROR! Unable to read the JSON index file from disk. Sorry...")
		fmt.Println("Please go to GitHub and open an issue.")
		os.Exit(0)
	}

	var allMedia []Media

	json.Unmarshal(src, &allMedia)

	fmt.Println("\nTrying to extract pictures and videos...")

	ch := make(chan bool, 10)
	mediaLock := sync.Mutex{}

	progressCount := len(allMedia)
	for _, media := range allMedia {
		if media.Status == "done" {
			progressCount--
		} else if !retryFlag && media.Status == "failed" {
			progressCount--
		}
	}

	progress := pb.New(progressCount)
	progress.ShowCounters = true
	progress.ShowTimeLeft = true
	progress.Start()

	fails := 0
	success := 0
	for i, media := range allMedia {
		if allMedia[i].Status == "done" {
			success += 1
			continue
		}

		if !retryFlag && allMedia[i].Status == "failed" {
			fails += 1
			continue
		}

		ch <- true

		go func(index int, media *Media) {
			fetchMedia(&client, media)
			mediaLock.Lock()
			allMedia[index] = *media
			if media.Status == "done" {
				success += 1
			} else {
				fails += 1
			}
			progress.Increment()
			mediaLock.Unlock()
			<-ch
		}(i, &media)

		if i > 0 && i%10 == 0 {
			mediaJson, _ := json.Marshal(allMedia)
			err = ioutil.WriteFile(indexPath, mediaJson, filePerm)

			if err != nil {
				fmt.Println("ERROR! Unable to write update JSON index file to disk. Sorry...")
				fmt.Println("Please go to GitHub and open an issue.")
				os.Exit(0)
			}
		}
	}

	mediaJson, _ := json.Marshal(allMedia)
	err = ioutil.WriteFile(indexPath, mediaJson, filePerm)

	if err != nil {
		fmt.Println("ERROR! Unable to write update JSON index file to disk. Sorry...")
		fmt.Println("Please go to GitHub and open an issue.")
		os.Exit(0)
	}

	progress.Finish()

	fmt.Println("Done trying to fetch all pictures and videos.")
	fmt.Println("Result:")
	fmt.Println("\tSuccess:", success)
	fmt.Println("\tFailed: ", fails)
}

func fetchMedia(client *minicrawler.Crawler, media *Media) {
	media.Retries += 1
	media.Status = "started"

	extension := strings.ToLower(media.Format)
	extension = strings.Replace(extension, "jpeg", "jpg", 1)
	filename := media.Id + "." + extension
	filePath := mediaPath + "/" + filename
	url := originalUrl.String() + media.Id

	out, err := os.Create(filePath)
	if err != nil {
		media.Status = "failed"
		out.Close()
		os.Remove(filePath)
		return
	}

	res, err := client.Client.Get(url)
	if err != nil || res.StatusCode != 200 {
		media.Status = "failed"
		out.Close()
		if res != nil {
			res.Body.Close()
		}
		os.Remove(filePath)
		return
	}

	n, err := io.Copy(out, res.Body)
	if err != nil {
		media.Status = "failed"
		out.Close()
		res.Body.Close()
		os.Remove(filePath)
		return
	}

	if n < 1000 {
		media.Status = "failed"
		out.Close()
		res.Body.Close()
		os.Remove(filePath)

	} else {
		media.Status = "done"
		out.Close()
		res.Body.Close()
	}
}

func printHelp() {
	fmt.Println("Currently you can only choose whether or not to retry failed fetches")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println(`./rescuelife -retry`)
	fmt.Println("")
}

func printStatus() {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Println("ERROR! Unable to find the JSON index file from disk. Sorry...")
		return
	}

	src, err := ioutil.ReadFile(indexPath)
	if err != nil {
		fmt.Println("ERROR! Unable to read the JSON index file from disk. Sorry...")
		return
	}

	var allMedia []Media

	json.Unmarshal(src, &allMedia)

	var failed, started, done, waiting int
	total := len(allMedia)
	for _, media := range allMedia {
		switch media.Status {
		case "done":
			done++
		case "started":
			started++
		case "failed":
			failed++
		default:
			waiting++
		}
	}

	fmt.Println("\nStatus for fetching")
	fmt.Println("-----------------------------")
	fmt.Println("Succeeded:", done)
	fmt.Println("Failed:   ", failed)
	fmt.Println("Fetching: ", started)
	fmt.Println("Waiting:  ", waiting)
	fmt.Println("Total:    ", total)
	fmt.Println("")
}

func getCredentials() (signinValues url.Values) {
	fmt.Println("\n---------------------------------------------------------------------------------------------------------------------")
	fmt.Println("Your email and password is needed in order to get a cookie, extract Access Token and to fetch your images and videos.")
	fmt.Println("Nothing will be stored or copied to any other server.")
	fmt.Println("---------------------------------------------------------------------------------------------------------------------\n")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Your email: ")
	email, _ := reader.ReadString('\n')
	email = strings.Trim(email, "\n")

	fmt.Print("Your password: ")
	bytePassword, _ := terminal.ReadPassword(0)
	password := strings.Trim(string(bytePassword), "\n")
	fmt.Println("\n")

	if email == "" || password == "" {
		fmt.Println("ERROR! Please provide email and password")
		os.Exit(0)
	}

	signinValues = url.Values{"email": {email}, "password": {password}}

	return
}
