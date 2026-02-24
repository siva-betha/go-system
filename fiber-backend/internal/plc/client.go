package plc

type Client interface {
    MachineID() string
    ReadSymbols(symbols []string) (map[string]interface{}, error)
    Close() error
}
