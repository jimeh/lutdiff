package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jimeh/lutdiff/lut"
	flag "github.com/spf13/pflag"
)

type options struct {
	start           string
	target          string
	output          string
	toneCurveIgnore []string
}

func main() {
	err := mainE(os.Args[1:])
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (*options, error) {
	opts := &options{}

	fs := flag.NewFlagSet("lutdiff", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`usage: lutdiff [flags] <start.json> <target.json>

lutdiff generates a LUT profile that transforms the colors in start.json to the
colors in target.json, essentially producing a "correction profile" that will
make footage shot with the start.json profile look like it was shot with the
target.json profile.`)

		fs.PrintDefaults()
	}

	fs.StringVarP(&opts.output, "output", "o", "", "output file")
	fs.StringArrayVarP(
		&opts.toneCurveIgnore,
		"ignore-tone-curve",
		"i",
		[]string{},
		"ignore tone curve values",
	)

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	fmt.Printf("fs.Args(): %#v\n", fs.Args())

	if fs.NArg() != 2 {
		return nil, errors.New("missing input files")
	}

	opts.start = fs.Arg(0)
	opts.target = fs.Arg(1)

	fmt.Printf("opts: %#v\n", opts)

	return opts, nil
}

func mainE(args []string) error {
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}

	start, err := readProfile(opts.start)
	if err != nil {
		return err
	}

	target, err := readProfile(opts.target)
	if err != nil {
		return err
	}

	profile, err := diffProfiles(start, target, opts)
	if err != nil {
		return err
	}

	if opts.output != "" {
		err = writeProfile(opts.output, profile)
		if err != nil {
			return err
		}
	} else {
		err = printProfile(profile)
		if err != nil {
			return err
		}
	}

	return nil
}

func readProfile(filename string) (*lut.Profile, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	p := &lut.Profile{}
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func diffProfiles(start, target *lut.Profile, opts *options) (*lut.Profile, error) {
	ignoreToneCurve := [][2]float64{}

	for _, s := range opts.toneCurveIgnore {
		var a, b float64
		_, err := fmt.Sscanf(s, "%f,%f", &a, &b)
		if err != nil {
			return nil, err
		}

		ignoreToneCurve = append(ignoreToneCurve, [2]float64{a, b})
	}

	n, err := start.Diff(target, ignoreToneCurve)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func printProfile(profile *lut.Profile) error {
	b, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}

func writeProfile(filename string, profile *lut.Profile) error {
	b, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, b, 0o644)
	if err != nil {
		return err
	}

	return nil
}