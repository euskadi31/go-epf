package epf

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

// Error type
var (
	ErrOutOfRange = errors.New("Out of range")
	ErrBadFormat  = errors.New("Bad format")
)

type stateType int

const (
	stateTypeFields stateType = iota
	stateTypePrimaryKey
	stateTypePrimaryKeyValue
	stateTypeTypes
	stateTypeTypesValue
	stateTypeExportMode
	stateTypeExportModeValue
	stateTypeComment
	stateTypeEnd
)

const (
	commentChar     rune = '#'
	recordSeparator rune = '\x02'
	fieldSeparator  rune = '\x01'
)

// Parser struct
type Parser struct {
	file     *os.File
	data     *bufio.Reader
	metadata *Metadata
}

// NewParser EPF
func NewParser(filename string) (*Parser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Parser{
		file: file,
		data: bufio.NewReader(file),
	}, nil
}

// Metadata of file
func (p *Parser) Metadata() *Metadata {
	if p.metadata == nil {
		p.loadHeaderInfo()
		p.loadFooterInfo()
	}

	return p.metadata
}

func (p Parser) convertType(value string, t string) (interface{}, error) {
	part := strings.Split(t, "(")

	if value == "" {
		return nil, nil
	}

	switch part[0] {
	case "BIGINT":
		return strconv.ParseInt(value, 10, 64)
	case "INTEGER":
		return strconv.Atoi(value)
	case "BOOLEAN":
		if value == "1" {
			return true, nil
		}

		return false, nil
	default:
		return value, nil
	}
}

func (p *Parser) isEndOfLine(r rune) bool {
	if r == recordSeparator {
		r, _, err := p.data.ReadRune()
		if err == io.EOF {
			return true
		}

		if r == '\n' {
			return true
		}

		p.data.UnreadRune()
	}

	return false
}

/*
1490173201020^A318519^AWilliam Boyce^A1^Ahttp://itunes.apple.com/artist/william-boyce/id318519?uo=5^A1^B
1490173201020^A320301^AK. Mills^A1^Ahttp://itunes.apple.com/artist/k-mills/id320301?uo=5^A1^B
1490173201020^A320355^AJ. Morris^A1^Ahttp://itunes.apple.com/artist/j-morris/id320355?uo=5^A1^B
1490173201020^A320417^AC. Levine^A1^Ahttp://itunes.apple.com/artist/c-levine/id320417?uo=5^A1^B
*/
// Read item
func (p *Parser) Read() (map[string]interface{}, error) {
	if p.metadata == nil {
		p.Metadata()
	}

	data := make(map[string]interface{})

	index := 0

	temp := ""

	isABeginningOfTheLine := true

	for {
		r, _, err := p.data.ReadRune()
		if err == io.EOF {
			return data, err
		}
		if err != nil {
			return data, err
		}

		// end of data, parse footer
		if isABeginningOfTheLine && r == commentChar {
			return data, io.EOF
		}

		if r == fieldSeparator || p.isEndOfLine(r) {
			if index > len(p.metadata.Fields)-1 {
				return data, ErrOutOfRange
			}

			field := p.metadata.Fields[index]
			t := p.metadata.Types[index]

			d, err := p.convertType(temp, t)
			if err != nil {
				return data, err
			}

			data[field] = d

			temp = ""

			if r == fieldSeparator {
				index++

				isABeginningOfTheLine = false

				continue
			}

			if r == recordSeparator {
				isABeginningOfTheLine = true

				return data, nil
			}
		}

		temp += string(r)

		isABeginningOfTheLine = false
	}
}

/*
#export_date^Aartist_id^Aname^Ais_actual_artist^Aview_url^Aartist_type_id^B
#primaryKey:artist_id^B
#dbTypes:BIGINT^AINTEGER^AVARCHAR(1000)^ABOOLEAN^AVARCHAR(1000)^AINTEGER^B
#exportMode:FULL^B
*/
func (p *Parser) loadHeaderInfo() error {
	p.metadata = &Metadata{
		Fields: []string{},
		Types:  []string{},
	}

	state := stateTypeFields

	temp := ""

	isNewLine := true

	for {
		r, _, err := p.data.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch state {
		case stateTypeFields:
			if isNewLine && r != commentChar {
				return ErrBadFormat
			} else if isNewLine && r == commentChar {
				isNewLine = false
				continue
			}

			if r == fieldSeparator || p.isEndOfLine(r) {
				p.metadata.Fields = append(p.metadata.Fields, temp)
				temp = ""

				if r == recordSeparator {
					isNewLine = true
					state = stateTypePrimaryKey
				}

				continue
			}

			temp += string(r)

		case stateTypePrimaryKey:
			if isNewLine && r != commentChar {
				return ErrBadFormat
			} else if isNewLine && r == commentChar {
				isNewLine = false

				continue
			}

			if r != ':' {
				continue
			}

			if r == ':' {
				state = stateTypePrimaryKeyValue
				continue
			}

		case stateTypePrimaryKeyValue:
			if r == fieldSeparator || p.isEndOfLine(r) {
				p.metadata.PrimaryKey = append(p.metadata.PrimaryKey, temp)
				temp = ""

				if r == recordSeparator {
					isNewLine = true
					state = stateTypeTypes
				}

				continue
			}

			temp += string(r)

		case stateTypeTypes:
			if isNewLine && r != commentChar {
				return ErrBadFormat
			} else if isNewLine && r == commentChar {
				isNewLine = false

				continue
			}

			if r != ':' {
				continue
			}

			if r == ':' {
				state = stateTypeTypesValue
				continue
			}

		case stateTypeTypesValue:
			if r == fieldSeparator || p.isEndOfLine(r) {
				p.metadata.Types = append(p.metadata.Types, temp)
				temp = ""

				if r == recordSeparator {
					isNewLine = true
					state = stateTypeExportMode
				}

				continue
			}

			temp += string(r)

		case stateTypeExportMode:
			if isNewLine && r != commentChar {
				return ErrBadFormat
			} else if isNewLine && r == commentChar {
				isNewLine = false

				continue
			}
			if r != ':' {
				continue
			}

			if r == ':' {
				state = stateTypeExportModeValue
				continue
			}

		case stateTypeExportModeValue:
			if p.isEndOfLine(r) {
				switch temp {
				case "FULL":
					p.metadata.ExportMode = ExportModeTypeFull
				default:
					p.metadata.ExportMode = ExportModeTypeIncremental
				}
				temp = ""
				isNewLine = true
				state = stateTypeComment

				continue
			}

			temp += string(r)

		case stateTypeComment:
			if isNewLine && r == commentChar {
				isNewLine = false
				continue
			}

			if isNewLine && r != commentChar {
				state = stateTypeEnd

				p.data.UnreadRune()

				return nil
			}

			if p.isEndOfLine(r) {
				isNewLine = true

				continue
			}

		default:
			return nil
		}
	}

	return nil
}

func (p *Parser) loadFooterInfo() error {
	//#recordsWritten:9981278

	stat, err := p.file.Stat()
	if err != nil {
		return err
	}

	buf := make([]byte, 28)
	n, err := p.file.ReadAt(buf, stat.Size()-int64(len(buf)))
	if err != nil {
		return err
	}
	buf = buf[:n]

	part := strings.Split(string(buf), ":")

	if len(part) != 2 {
		return errors.New("Cannot read footer info")
	}

	recordsWritten := strings.Trim(part[1], "\x02\n")

	value, err := strconv.Atoi(recordsWritten)
	if err != nil {
		return err
	}

	p.metadata.TotalItems = value

	return nil
}

// Close Parser
func (p *Parser) Close() {
	p.file.Close()
}
