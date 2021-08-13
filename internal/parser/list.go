package parser

import (
	"fmt"
)

type List struct {
	Commands []*Command
}

func (l *List) Items() (map[string]*Item, error) {
	var result = make(map[string]*Item)

	for _, c := range l.Commands {
		items := result
		var item *Item
		var ok bool

		for _, path := range c.Path {
			item, ok = items[path]

			if !ok {
				item = new(Item)
				item.Level = LevelTypePath
				item.Children = make(map[string]*Item)
				item.Payload = &Payload{}

				items[path] = item
			}

			items = item.Children
		}

		item, ok = items[c.Name]

		if !ok {
			item = new(Item)
			item.Level = LevelTypeCommand
			item.Children = make(map[string]*Item)
			item.Payload = c

			c.unnamedFlags = make(map[uint]*Flag)

			items[c.Name] = item
		}

		items = item.Children

		for flagName, flag := range c.Flags {
			item, ok = items[flagName]

			if ok {
				continue
			}

			if flag.Number > 0 {
				if !flag.Mandatory {
					return nil, fmt.Errorf("flag %v is unnamed and not mandatory", flagName)
				}

				if _, ok := c.unnamedFlags[flag.Number]; ok {
					return nil, fmt.Errorf("number %v is already set", flag.Number)
				}

				c.unnamedFlags[flag.Number] = flag
			}

			item = new(Item)
			item.Level = LevelTypeFlag

			item.Payload = flag

			if flag.Mandatory {
				c.MandatoryFlags = append(c.MandatoryFlags, flagName)
			}

			items[flagName] = item
		}

		for optionName, option := range c.Options {
			item, ok = items[optionName]

			if ok {
				continue
			}

			item = new(Item)
			item.Level = LevelTypeOption
			item.Payload = option

			items[optionName] = item
		}

		for i := 1; i <= len(c.unnamedFlags); i++ {
			_, ok := c.unnamedFlags[uint(i)]

			if !ok {
				return nil, fmt.Errorf("unnamed flags are ordered wrong")
			}
		}
	}

	return result, nil
}
