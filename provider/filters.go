package provider

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/geoidx"
	"github.com/vatsimnerd/lee/lexer"
	"github.com/vatsimnerd/lee/parser"
	"github.com/vatsimnerd/simwatch-providers/merged"
)

func airportFilter(includeUncontrolled bool) geoidx.Filter {
	if !includeUncontrolled {
		return func(obj *geoidx.Object) bool {
			if arpt, ok := obj.Value().(merged.Airport); ok {
				return arpt.IsControlled()
			}
			return true
		}
	}
	return nil
}

func pilotFilter(query string) (geoidx.Filter, error) {
	log := logrus.WithFields(logrus.Fields{
		"func":  "planeFilter",
		"query": query,
	})

	t1 := time.Now()

	tokens, err := lexer.Tokenize(query, true)
	if err != nil {
		return nil, err
	}

	expr, err := parser.Parse[*geoidx.Object](tokens)
	if err != nil {
		return nil, err
	}

	err = expr.Compile(func(c *parser.Condition[*geoidx.Object]) (parser.Matcher[*geoidx.Object], error) {
		log.WithField("condition", c.String()).Debug("compiling condition")

		switch c.Identifier.Name {
		case "aircraft":
			if !c.Value.IsString() {
				return nil, fmt.Errorf("missing string value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetStringValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return pilot.FlightPlan.Aircraft == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return pilot.FlightPlan.Aircraft != value
					}
					return false
				}, nil
			case parser.Matches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return re.MatchString(pilot.FlightPlan.Aircraft)
					}
					return false
				}, nil
			case parser.NotMatches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return !re.MatchString(pilot.FlightPlan.Aircraft)
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "departure":
			if !c.Value.IsString() {
				return nil, fmt.Errorf("missing string value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetStringValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return pilot.FlightPlan.Departure == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return pilot.FlightPlan.Departure != value
					}
					return false
				}, nil
			case parser.Matches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return re.MatchString(pilot.FlightPlan.Departure)
					}
					return false
				}, nil
			case parser.NotMatches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return !re.MatchString(pilot.FlightPlan.Departure)
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "arrival":
			if !c.Value.IsString() {
				return nil, fmt.Errorf("missing string value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetStringValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return pilot.FlightPlan.Arrival == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return pilot.FlightPlan.Arrival != value
					}
					return false
				}, nil
			case parser.Matches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return re.MatchString(pilot.FlightPlan.Arrival)
					}
					return false
				}, nil
			case parser.NotMatches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						if pilot.FlightPlan == nil {
							return false
						}
						return !re.MatchString(pilot.FlightPlan.Arrival)
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "callsign":
			if !c.Value.IsString() {
				return nil, fmt.Errorf("missing string value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetStringValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Callsign == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Callsign != value
					}
					return false
				}, nil
			case parser.Matches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return re.MatchString(pilot.Callsign)
					}
					return false
				}, nil
			case parser.NotMatches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return !re.MatchString(pilot.Callsign)
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "name":
			if !c.Value.IsString() {
				return nil, fmt.Errorf("missing string value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetStringValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Name == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Name != value
					}
					return false
				}, nil
			case parser.Matches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return re.MatchString(pilot.Name)
					}
					return false
				}, nil
			case parser.NotMatches:
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling expression %s: %v", value, err)
				}

				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return !re.MatchString(pilot.Name)
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "alt":
			if !c.Value.IsFloat() {
				return nil, fmt.Errorf("missing numeric value for %s", c.Identifier.Name)
			}
			value := int(c.Value.MustGetFloatValue())

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Altitude == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Altitude != value
					}
					return false
				}, nil
			case parser.Less:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Altitude < value
					}
					return false
				}, nil
			case parser.LessOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Altitude <= value
					}
					return false
				}, nil
			case parser.Greater:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Altitude > value
					}
					return false
				}, nil
			case parser.GreaterOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Altitude >= value
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "gs":
			if !c.Value.IsFloat() {
				return nil, fmt.Errorf("missing numeric value for %s", c.Identifier.Name)
			}
			value := int(c.Value.MustGetFloatValue())

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Groundspeed == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Groundspeed != value
					}
					return false
				}, nil
			case parser.Less:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Groundspeed < value
					}
					return false
				}, nil
			case parser.LessOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Groundspeed <= value
					}
					return false
				}, nil
			case parser.Greater:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Groundspeed > value
					}
					return false
				}, nil
			case parser.GreaterOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Groundspeed >= value
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "lat":
			if !c.Value.IsFloat() {
				return nil, fmt.Errorf("missing numeric value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetFloatValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Latitude == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Latitude != value
					}
					return false
				}, nil
			case parser.Less:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Latitude < value
					}
					return false
				}, nil
			case parser.LessOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Latitude <= value
					}
					return false
				}, nil
			case parser.Greater:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Latitude > value
					}
					return false
				}, nil
			case parser.GreaterOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Latitude >= value
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "lng":
			if !c.Value.IsFloat() {
				return nil, fmt.Errorf("missing numeric value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetFloatValue()

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Longitude == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Longitude != value
					}
					return false
				}, nil
			case parser.Less:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Longitude < value
					}
					return false
				}, nil
			case parser.LessOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Longitude <= value
					}
					return false
				}, nil
			case parser.Greater:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Longitude > value
					}
					return false
				}, nil
			case parser.GreaterOrEqual:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.Longitude >= value
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		case "rules":
			if !c.Value.IsString() {
				return nil, fmt.Errorf("missing string value for %s", c.Identifier.Name)
			}
			value := c.Value.MustGetStringValue()
			value = strings.ToLower(value)
			if value != "i" && value != "v" && value != "vfr" && value != "ifr" {
				return nil, fmt.Errorf("invalid value for %s, expected I/V/IFR/VFR", c.Identifier.Name)
			}
			value = strings.ToUpper(value[0:1])

			switch c.Operator.Type {
			case parser.Equals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.FlightPlan != nil && pilot.FlightPlan.FlightRules == value
					}
					return false
				}, nil
			case parser.NotEquals:
				return func(obj *geoidx.Object) bool {
					if pilot, ok := obj.Value().(merged.Pilot); ok {
						return pilot.FlightPlan != nil && pilot.FlightPlan.FlightRules != value
					}
					return false
				}, nil
			default:
				return nil, fmt.Errorf("invalid operator %s for %s", c.Operator.Type, c.Identifier.Name)
			}
		default:
			return nil, fmt.Errorf("field %s is invalid or not supported yet", c.Identifier.Name)
		}
	})

	if err != nil {
		return nil, err
	}

	t2 := time.Now()
	log.WithField("time", t2.Sub(t1).String()).Debug("expression compiled")

	return expr.Evaluate, nil
}
