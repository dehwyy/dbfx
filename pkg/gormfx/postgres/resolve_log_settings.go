package postgres

import "os"

func resolveLogSettings(opts Opts, get func(string) string) (logSettings, error) {
	level, err := resolveLogLevel(
		opts.LogLevel,
		get,
	)
	if err != nil {
		return logSettings{}, err
	}

	slow, err := resolveLogSlowThreshold(
		opts.LogSlowThreshold,
		get,
	)
	if err != nil {
		return logSettings{}, err
	}

	output := opts.LogOutput
	if output == nil {
		output = os.Stdout
	}

	return logSettings{
		level:  level,
		slow:   slow,
		output: output,
	}, nil
}
