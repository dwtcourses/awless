package template

import (
	"fmt"
	"strings"

	"github.com/wallix/awless/template/internal/ast"
)

func (te *Template) Revert() (*Template, error) {
	var lines []string
	cmdsReverseIterator := te.CommandNodesReverseIterator()
	for i, cmd := range cmdsReverseIterator {
		notLastCommand := (i != len(cmdsReverseIterator)-1)
		if isRevertible(cmd) {
			var revertAction string
			var params []string

			switch cmd.Action {
			case "create", "copy":
				revertAction = "delete"
			case "start":
				revertAction = "stop"
			case "stop":
				revertAction = "start"
			case "detach":
				revertAction = "attach"
			case "attach":
				revertAction = "detach"
			case "delete":
				revertAction = "create"
			case "update":
				revertAction = "update"
			}

			switch cmd.Action {
			case "attach":
				switch cmd.Entity {
				case "routetable", "elasticip":
					params = append(params, fmt.Sprintf("association=%s", quoteParamIfNeeded(cmd.CmdResult)))
				case "instance":
					for k, v := range cmd.Params {
						if k == "port" {
							continue
						}
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				case "containertask":
					params = append(params, fmt.Sprintf("name=%s", cmd.Params["name"].String()))
					params = append(params, fmt.Sprintf("container-name=%s", cmd.Params["container-name"].String()))
				case "networkinterface":
					params = append(params, fmt.Sprintf("attachment=%s", quoteParamIfNeeded(cmd.CmdResult)))
				case "mfadevice":
					params = append(params, fmt.Sprintf("id=%s", cmd.Params["id"].String()))
					params = append(params, fmt.Sprintf("user=%s", cmd.Params["user"].String()))
				default:
					for k, v := range cmd.Params {
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				}
			case "start", "stop", "detach":
				switch {
				case cmd.Entity == "routetable":
					params = append(params, fmt.Sprintf("association=%s", quoteParamIfNeeded(cmd.CmdResult)))
				case cmd.Entity == "volume" && cmd.Action == "detach":
					for k, v := range cmd.Params {
						if k == "force" {
							continue
						}
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				case cmd.Entity == "containertask":
					params = append(params, fmt.Sprintf("cluster=%s", cmd.Params["cluster"].String()))
					params = append(params, fmt.Sprintf("type=%s", cmd.Params["type"].String()))
					switch fmt.Sprint(cmd.Params["type"]) {
					case "service":
						params = append(params, fmt.Sprintf("deployment-name=%s", cmd.Params["deployment-name"].String()))
					case "task":
						params = append(params, fmt.Sprintf("run-arn=%s", quoteParamIfNeeded(cmd.CmdResult)))
					default:
						return nil, fmt.Errorf("start containertask with type '%v' can not be reverted", cmd.Params["deployment-name"].String())
					}
				default:
					for k, v := range cmd.Params {
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				}
			case "create":
				switch cmd.Entity {
				case "tag":
					for k, v := range cmd.Params {
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				case "record":
					for k, v := range cmd.Params {
						if k == "comment" {
							continue
						}
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				case "route":
					for k, v := range cmd.Params {
						if k == "gateway" {
							continue
						}
						params = append(params, fmt.Sprintf("%s=%v", k, v.String()))
					}
				case "database":
					params = append(params, fmt.Sprintf("id=%s", quoteParamIfNeeded(cmd.CmdResult)))
					params = append(params, "skip-snapshot=true")
				case "certificate":
					params = append(params, fmt.Sprintf("arn=%s", quoteParamIfNeeded(cmd.CmdResult)))
				case "policy":
					params = append(params, fmt.Sprintf("arn=%s", quoteParamIfNeeded(cmd.CmdResult)))
					params = append(params, "all-versions=true")
				case "queue":
					params = append(params, fmt.Sprintf("url=%s", quoteParamIfNeeded(cmd.CmdResult)))
				case "s3object":
					params = append(params, fmt.Sprintf("name=%s", quoteParamIfNeeded(cmd.CmdResult)))
					params = append(params, fmt.Sprintf("bucket=%s", cmd.Params["bucket"].String()))
				case "role", "group", "user", "stack", "instanceprofile", "repository":
					params = append(params, fmt.Sprintf("name=%s", cmd.Params["name"].String()))
				case "accesskey":
					params = append(params, fmt.Sprintf("id=%s", quoteParamIfNeeded(cmd.CmdResult)))
					params = append(params, fmt.Sprintf("user=%s", cmd.Params["user"].String()))
				case "appscalingtarget":
					params = append(params, fmt.Sprintf("dimension=%s", cmd.Params["dimension"].String()))
					params = append(params, fmt.Sprintf("resource=%s", cmd.Params["resource"].String()))
					params = append(params, fmt.Sprintf("service-namespace=%s", cmd.Params["service-namespace"].String()))
				case "appscalingpolicy":
					params = append(params, fmt.Sprintf("dimension=%s", cmd.Params["dimension"].String()))
					params = append(params, fmt.Sprintf("name=%s", cmd.Params["name"].String()))
					params = append(params, fmt.Sprintf("resource=%s", cmd.Params["resource"].String()))
					params = append(params, fmt.Sprintf("service-namespace=%s", cmd.Params["service-namespace"].String()))
				case "loginprofile":
					params = append(params, fmt.Sprintf("username=%s", cmd.Params["username"].String()))
				case "bucket", "launchconfiguration", "scalinggroup", "alarm", "dbsubnetgroup", "keypair":
					params = append(params, fmt.Sprintf("name=%s", quoteParamIfNeeded(cmd.CmdResult)))
					if cmd.Entity == "scalinggroup" {
						params = append(params, "force=true")
					}
				default:
					params = append(params, fmt.Sprintf("id=%s", quoteParamIfNeeded(cmd.CmdResult)))
				}
			case "delete":
				switch cmd.Entity {
				case "record":
					for k, v := range cmd.Params {
						params = append(params, fmt.Sprintf("%s=%v", k, quoteParamIfNeeded(v)))
					}
				case "instanceprofile":
					params = append(params, fmt.Sprintf("name=%s", cmd.Params["name"].String()))
				}
			case "copy":
				switch cmd.Entity {
				case "image":
					params = append(params, fmt.Sprintf("id=%s", quoteParamIfNeeded(cmd.CmdResult)))
					params = append(params, "delete-snapshots=true")
				default:
					params = append(params, fmt.Sprintf("id=%s", quoteParamIfNeeded(cmd.CmdResult)))
				}
			case "update":
				switch cmd.Entity {
				case "securitygroup":
					for k, v := range cmd.Params {
						if k == "inbound" || k == "outbound" {
							if fmt.Sprint(v) == "authorize" {
								params = append(params, fmt.Sprintf("%s=revoke", k))
							} else if fmt.Sprint(v) == "revoke" {
								params = append(params, fmt.Sprintf("%s=authorize", k))
							}
							continue
						}
						params = append(params, fmt.Sprintf("%s=%v", k, quoteParamIfNeeded(v)))
					}
				}
			}

			// Prechecks
			if cmd.Action == "create" && cmd.Entity == "securitygroup" {
				lines = append(lines, fmt.Sprintf("check securitygroup id=%s state=unused timeout=300", quoteParamIfNeeded(cmd.CmdResult)))
			}
			if cmd.Action == "create" && cmd.Entity == "scalinggroup" {
				lines = append(lines, fmt.Sprintf("update scalinggroup name=%s max-size=0 min-size=0", quoteParamIfNeeded(cmd.CmdResult)))
				lines = append(lines, fmt.Sprintf("check scalinggroup count=0 name=%s timeout=600", quoteParamIfNeeded(cmd.CmdResult)))
			}
			if cmd.Action == "start" && cmd.Entity == "instance" {
				switch vv := cmd.ToDriverParams()["ids"].(type) {
				case string:
					lines = append(lines, fmt.Sprintf("check instance id=%s state=running timeout=180", quoteParamIfNeeded(vv)))
				case []interface{}:
					for _, s := range vv {
						lines = append(lines, fmt.Sprintf("check instance id=%v state=running timeout=180", quoteParamIfNeeded(s)))
					}
				}
			}
			if cmd.Action == "stop" && cmd.Entity == "instance" {
				switch vv := cmd.ToDriverParams()["ids"].(type) {
				case string:
					lines = append(lines, fmt.Sprintf("check instance id=%s state=stopped timeout=180", quoteParamIfNeeded(vv)))
				case []interface{}:
					for _, s := range vv {
						lines = append(lines, fmt.Sprintf("check instance id=%v state=stopped timeout=180", quoteParamIfNeeded(s)))
					}
				}
			}
			if cmd.Action == "start" && cmd.Entity == "containertask" && fmt.Sprint(cmd.Params["type"]) == "service" {
				lines = append(lines, fmt.Sprintf("update containertask cluster=%s deployment-name=%s desired-count=0", cmd.Params["cluster"].String(), cmd.Params["deployment-name"].String()))
			}

			lines = append(lines, fmt.Sprintf("%s %s %s", revertAction, cmd.Entity, strings.Join(params, " ")))

			// Postchecks
			if notLastCommand {
				if cmd.Action == "create" && cmd.Entity == "instance" {
					lines = append(lines, fmt.Sprintf("check instance id=%s state=terminated timeout=180", quoteParamIfNeeded(cmd.CmdResult)))
				}
				if cmd.Action == "create" && cmd.Entity == "database" {
					lines = append(lines, fmt.Sprintf("check database id=%s state=not-found timeout=900", quoteParamIfNeeded(cmd.CmdResult)))
				}
				if cmd.Action == "create" && cmd.Entity == "loadbalancer" {
					lines = append(lines, fmt.Sprintf("check loadbalancer id=%s state=not-found timeout=180", quoteParamIfNeeded(cmd.CmdResult)))
				}
				if cmd.Action == "attach" && cmd.Entity == "volume" {
					lines = append(lines, fmt.Sprintf("check volume id=%s state=available timeout=180", cmd.Params["id"].String()))
				}
				if cmd.Action == "create" && cmd.Entity == "natgateway" {
					lines = append(lines, fmt.Sprintf("check natgateway id=%s state=deleted timeout=180", quoteParamIfNeeded(cmd.CmdResult)))
				}
			}
		}
	}

	text := strings.Join(lines, "\n")
	tpl, err := Parse(text)
	if err != nil {
		return nil, fmt.Errorf("revert: \n%s\n%s", text, err)
	}

	return tpl, nil
}

func IsRevertible(t *Template) bool {
	revertible := false
	t.visitCommandNodes(func(cmd *ast.CommandNode) {
		if isRevertible(cmd) {
			revertible = true
		}
	})
	return revertible
}

func isRevertible(cmd *ast.CommandNode) bool {
	if cmd.CmdErr != nil {
		return false
	}

	if cmd.Action == "check" {
		return false
	}

	if cmd.Action == "detach" && cmd.Entity == "routetable" {
		return false
	}

	if cmd.Entity == "record" && (cmd.Action == "create" || cmd.Action == "delete") {
		return true
	}

	if cmd.Entity == "instanceprofile" && (cmd.Action == "create" || cmd.Action == "delete") {
		return true
	}

	if cmd.Entity == "alarm" && (cmd.Action == "start" || cmd.Action == "stop") {
		return true
	}

	if cmd.Entity == "database" && (cmd.Action == "start" || cmd.Action == "stop") {
		return true
	}

	if cmd.Entity == "containertask" && cmd.Action == "start" {
		t, ok := cmd.ToDriverParams()["type"].(string)
		return ok && (t == "service" || t == "task")
	}

	if cmd.Entity == "container" && cmd.Action == "create" {
		return true
	}

	if cmd.Entity == "appscalingtarget" && cmd.Action == "create" {
		return true
	}

	if cmd.Entity == "securitygroup" && cmd.Action == "update" {
		return true
	}

	if cmd.Entity == "appscalingpolicy" && cmd.Action == "create" {
		return true
	}

	if v, ok := cmd.CmdResult.(string); ok && v != "" {
		if cmd.Action == "create" || cmd.Action == "start" || cmd.Action == "stop" || cmd.Action == "copy" {
			return true
		}
	}

	return cmd.Action == "attach" || cmd.Action == "detach" || cmd.Action == "check" ||
		(cmd.Action == "create" && cmd.Entity == "tag") || (cmd.Action == "create" && cmd.Entity == "route")
}

func quoteParamIfNeeded(param interface{}) string {
	input := fmt.Sprint(param)
	if ast.SimpleStringValue.MatchString(input) {
		return input
	} else {
		if strings.ContainsRune(input, '\'') {
			return "\"" + input + "\""
		} else {
			return "'" + input + "'"
		}
	}
}
