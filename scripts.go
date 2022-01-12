package dsf

const (
	setNXOrGet = `
		if redis.pcall('SET', KEYS[1], ARGV[1], 'EX', ARGV[2], 'NX')
		then
			return {1, ""}
		else
			return {0, redis.pcall('GET', KEYS[1])}
		end
	`
)
