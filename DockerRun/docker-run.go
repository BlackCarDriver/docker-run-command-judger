package DockerRun

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/* the flags that can be recognized in docker run command:
-a, --attach
-c, --cpu-shares
-e, --env
-h, --hostname
-l, --label
-p, --publish
-u, --user
-v, --volume
--name
-w, --workdir
--link
-m, --memory
-i, --interactive
-d, --detach
--rm
-t, --tty
-P, --publish-all
*/

//the following is all flags allowled to used
var (
	//SimpleFlagList is those flag start with '-', such as -p -v -i
	SimpleFlagList = []string{"p", "P", "v", "i", "d", "t", "w", "u", "a", "c", "e", "h", "l", "u", "m"}
	//multiFlagList is those flag start with '--', such as --volume, --link
	MultiFlagList = []string{"publish-all", "tty", "rm", "detach", "interactive", "link", "workdir",
		"name", "volume", "user", "label", "hostname", "env", "cpu-shares", "attach", "memory", "network"}
	//NoArgFlagList is those flag attach with no arguments, such as -d, -i, --rm, note that P and p is different!
	NoArgFlagList = []string{"i", "interactive", "t", "tty", "d", "detach", "rm", "P", "publish-all"}
)

//the model to simulate a container property
type MockContainer struct {
	Images        string
	Command       string
	Arg           []string
	Port          map[string]string
	Volume        map[string]string
	Env           map[string]string
	Label         []string
	CpuShare      int
	Memory        int
	HostName      string
	ContainerName string
	User          string
	WorkDir       string
	NetWork       string
	IsRemove      bool
	IsDetach      bool
	IsTTY         bool
	IsInteractive bool
	IsPublishAll  bool
	Attach        []string
	Link          []string
}

//create an MockContainer according to a docker run command, return error if it command have a worng syntax
func NewMockContainer(dockerCmd string) (model MockContainer, err error) {
	model.Port = make(map[string]string)
	model.Volume = make(map[string]string)
	model.Env = make(map[string]string)
	cmdArray := splitCommand(dockerCmd)
	result := model.BasicCheck(cmdArray)
	if result == "" {
		return model, nil
	}
	return model, fmt.Errorf(result)
}

//check the basic syntax of a docker run command,
//return the fall reason or return a empty string if the command is accpeted
//synatax: docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
func (this *MockContainer) BasicCheck(cmd []string) string {
	var err error
	if len(cmd) == 0 {
		return "Receive empty command!"
	}
	if len(cmd) < 2 {
		return "Requires at least two element!"
	}
	if cmd[0] != "docker" {
		return "Not a docker command!"
	}
	if cmd[1] != "run" {
		return "Not a run command!"
	}
	//begain to explain option part
	nowAt := 1
	for {
		nowAt++
		if nowAt >= len(cmd) {
			break
		}
		tflag := cmd[nowAt]
		arg := ""
		if strings.HasPrefix(tflag, "--") { //scuh as --rm --volume
			flag := strings.TrimLeft(tflag, "--")
			if index := strings.Index(flag, "="); index > 0 { //have a '=', such as --volume=test --rm=true
				if index+1 == len(flag) { //no argument following '=', such as 'rm='
					return fmt.Sprintf("Unexpect flag: %s", tflag)
				}
				arg = flag[index+1:]
				flag = flag[0:index]
			}
			if !findInArray(MultiFlagList, flag) {
				return fmt.Sprintf("Unknown flag: %s", tflag)
			}
			if findInArray(NoArgFlagList, flag) { //don't need argument by default, such as --rm --tty
				if arg == "" || arg == "true" {
					err = this.HandleFlag(flag)
					if err != nil {
						return fmt.Sprint(err)
					}
				} else if arg != "false" {
					return fmt.Sprintf("Unexpect flag and argument: %s=%s", flag, arg)
				}
			} else { //need a argument, such as --name
				if arg != "" { //--name=hello
					err = this.HandleArgument(flag, arg)
				} else { //--name hello
					nowAt++
					if len(cmd) <= nowAt {
						return fmt.Sprintf("Not enough of argument after %s", tflag)
					}
					err = this.HandleArgument(flag, cmd[nowAt])
				}
				if err != nil {
					return fmt.Sprint(err)
				}
			}
		} else if strings.HasPrefix(tflag, "-") { //such as -p -d
			flags := strings.TrimLeft(tflag, "-")
			for i := 0; i < len(flags); i++ {
				arg := ""
				flag := flags[i : i+1]
				if !findInArray(SimpleFlagList, flag) {
					return fmt.Sprintf("unknown shorthand flag: %s in %s ", flag, flags[i+1:])
				}
				if findInArray(NoArgFlagList, flag) { //do not have argument by default, like -p -t
					if i+2 < len(flags) && flags[i+1] == '=' { //-t=true
						arg = flags[i+2:]
						i = len(flags)
					}
					if arg == "" || arg == "true" {
						err = this.HandleFlag(flag)
						if err != nil {
							return fmt.Sprint(err)
						}
					} else if arg != "false" {
						return fmt.Sprintf("Unexpect argument: %s=%s", flag, arg)
					}
				} else {
					if i+1 < len(flags) { //such as -ip8080:8080 or -ip=8080:8080
						arg = trimStr(flags[i+1:])
						if strings.HasPrefix(arg, "=") {
							arg = arg[1:]
						}
						err = this.HandleArgument(flag, arg)
						if err != nil {
							return fmt.Sprint(err)
						}
						break
					} else { //such as -ip 8080:8080
						nowAt++
						if nowAt >= len(cmd) {
							return fmt.Sprintf("Not enough of argument after -%s", flag)
						}
						arg = cmd[nowAt]
						err = this.HandleArgument(flag, arg)
						if err != nil {
							return fmt.Sprint(err)
						}
					}
				}
			}
		} else { //not a flag
			break
		}
	}
	//begain to read images name
	if nowAt >= len(cmd) {
		return "Can't find images name from given command!"
	}
	tImagesName := cmd[nowAt]
	if isImagesName(tImagesName) {
		if strings.Index(tImagesName, ":") < 0 {
			tImagesName += ":latest"
		}
		this.Images = tImagesName
	} else {
		return fmt.Sprintf("Images name %s not legal!", tImagesName)
	}
	nowAt++
	//begain to read Command and Arguments
	if nowAt >= len(cmd) { //no command
		return ""
	}
	this.Command = cmd[nowAt]
	nowAt++
	if nowAt >= len(cmd) { //no argument
		return ""
	}
	this.Arg = cmd[nowAt:]
	return ""
}

//Setting up the property of a container according to the flag and argument
//if the format of arguments not right it will return error
func (this *MockContainer) HandleArgument(flag, arg string) error {
	switch flag {
	case "p", "publish":
		arg = trimStr(arg)
		if !isPortArg(arg) {
			return fmt.Errorf("invalid publish opts format (should be port1:port2 but got '%s').", arg)
		}
		ports := strings.Split(arg, ":")
		_, have := this.Port[ports[0]]
		if have {
			return fmt.Errorf("Port is already allocated: %s", ports[0])
		}
		this.Port[ports[0]] = ports[1]
	case "c", "cpu-shares":
		share, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("%s need a number, but got: %s", flag, arg)
		}
		if share < 2 || share > 262144 {
			return fmt.Errorf("The allowed cpu-shares is from 2 to 262144")
		}
		this.CpuShare = share
	case "v", "volume":
		arg = trimStr(arg)
		if !isDirPath(arg) {
			return fmt.Errorf("Invalid volume argument: %s", arg)
		}
		paths := strings.Split(arg, ":")
		localPath := paths[0]
		conPart := paths[1]
		if strings.Index(localPath, "$(pwd)") >= 0 {
			localPath = strings.Replace(localPath, "$(pwd)", "$PWD", 1)
		}
		localPath = strings.TrimRight(localPath, "\\/")
		conPart = strings.TrimRight(conPart, "\\/")
		for _, v := range this.Volume {
			if v == conPart {
				return fmt.Errorf("Duplicate mount point: %s", conPart)
			}
		}
		this.Volume[localPath] = conPart

	case "name":
		arg = trimStr(arg)
		if !isContainerName(arg) {
			return fmt.Errorf("Invalid container name (%s), only [a-zA-Z0-9][a-zA-Z0-9_.-] are allowed.", arg)
		}
		this.ContainerName = arg
	case "network":
		this.NetWork = arg
	case "u", "user":
		this.User = arg
	case "w", "workdir":
		arg = trimStr(arg)
		if !isWorkDir(arg) {
			return fmt.Errorf("Invali workdir: %s", arg)
		}
		this.WorkDir = arg
	case "h", "hostname":
		this.HostName = strings.Trim(arg, "\"")
	case "e", "env":
		arg = trimStr(arg)
		this.Env[flag] = arg
	case "a", "attach":
		arg = trimStr(arg)
		if !isAttach(arg) {
			return fmt.Errorf("Invalid argument '%s' for -a, --attach", arg)
		}
		this.Attach = append(this.Attach, arg)
	case "l", "label":
		this.Label = append(this.Label, arg)
	case "link":
		this.Link = append(this.Link, arg)
	case "m", "memory":
		if !isMemory(arg) {
			return fmt.Errorf("Invalid memory argument: %s", arg)
		}
		tindex := strings.IndexAny(arg, "bBkKmMgG")
		if tindex < 0 {
			return fmt.Errorf("Invalid memory argument: %s", arg)
		}
		numStr := arg[:tindex]
		MetaStr := arg[tindex:]
		tnum, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("Invalid memory argument: %s", arg)
		}
		if len(MetaStr) == 0 {
			return fmt.Errorf("Invalid memory argument: %s", arg)
		}
		MetaStr = strings.ToLower(MetaStr)
		switch MetaStr[0] {
		case 'b':
			this.Memory = tnum >> 20
		case 'k':
			this.Memory = tnum >> 10
		case 'm':
			this.Memory = tnum
		case 'g':
			this.Memory = tnum << 10
		default:
			return fmt.Errorf("Invalid memory argument: %s", arg)
		}
	default:
		return fmt.Errorf("Invalid flag: --%s", flag)
	}
	return nil
}

//Setting up the property of a container according to the flag that without argument
//return error only if the flag is not exist
func (this *MockContainer) HandleFlag(flag string) error {
	switch flag {
	case "P", "publish-all":
		this.IsPublishAll = true
	case "t", "tty":
		this.IsTTY = true
	case "i", "interactive":
		this.IsInteractive = true
	case "rm":
		this.IsRemove = true
	case "d", "detach":
		this.IsDetach = true
	default:
		return fmt.Errorf("unknown shorthand flag: '%s'", flag)
	}
	return nil
}

//printf the property that have been changed of a conatiner
func (this *MockContainer) Printf() {
	psic := func(tag, str string) {
		if str != "" {
			fmt.Printf("%s  :  %s \n", tag, str)
		}
	}
	pssic := func(tag string, strs []string) {
		if len(strs) > 0 {
			fmt.Printf("%s  :  %v \n", tag, strs)
		}
	}
	pbic := func(tag string, b bool) {
		if b {
			fmt.Printf("%s  :  %v \n", tag, b)
		}
	}
	pmic := func(tag string, m map[string]string) {
		if len(m) > 0 {
			fmt.Printf("%s  :  %v \n", tag, m)
		}
	}
	psic("Images", this.Images)
	psic("Command", this.Command)
	psic("HostName", this.HostName)
	psic("ContainerName", this.ContainerName)
	psic("User", this.User)
	psic("WorkDir", this.WorkDir)
	psic("NetWork", this.NetWork)
	pbic("IsRemove", this.IsRemove)
	pbic("IsDetach", this.IsDetach)
	pbic("IsTTY", this.IsTTY)
	pbic("IsInteractive", this.IsInteractive)
	pbic("IsPublishAll", this.IsPublishAll)
	pssic("Arg", this.Arg)
	pssic("Attach", this.Attach)
	pssic("Link", this.Link)
	pssic("Label", this.Label)
	pmic("Port", this.Port)
	pmic("Volume", this.Volume)
	pmic("Env", this.Env)
	if this.Memory != 0 {
		fmt.Println("Memory  :  ", this.Memory)
	}
}

//judge if the property of a container is right by compared to the answer
//return a string to describe the mistake or a null string if it command is accepted
//note that here we have some config do not check: Env[], Label[], Attach[], Link[]
func Judge(test, ans *MockContainer) string {
	if test == nil {
		return "Given pointer of test is null"
	}
	if ans == nil {
		return "Given pointer of ans is null!"
	}
	if ans.IsTTY && !test.IsTTY {
		return "Not found -t or --tty."
	}
	if ans.IsDetach && !test.IsDetach {
		return "Not found -d or --detach"
	}
	if ans.IsRemove && !test.IsRemove {
		return "Not found --rm"
	}
	if ans.IsInteractive && !test.IsInteractive {
		return "not found -i or --interactive"
	}
	if ans.IsPublishAll && !test.IsPublishAll {
		return "not found -P or --publish-all"
	}
	if ans.WorkDir != "" && test.WorkDir != ans.WorkDir {
		return fmt.Sprintf("WorkDir not right, expect '%s' but got '%s'.", ans.WorkDir, test.WorkDir)
	}
	if ans.ContainerName != "" && test.ContainerName != ans.ContainerName {
		return fmt.Sprintf("ContainerName not right, expect '%s' but got '%s'.", ans.ContainerName, test.ContainerName)
	}
	if ans.User != "" && test.User != ans.User {
		return fmt.Sprintf("User not right, expect '%s' but got '%s'.", ans.User, test.User)
	}
	if ans.HostName != "" && test.HostName != ans.HostName {
		return fmt.Sprintf("HostName not right, expect '%s' but got '%s'.", ans.HostName, test.HostName)
	}
	if ans.CpuShare != test.CpuShare {
		return fmt.Sprintf("CpuShare not right, expect %d but got %d", ans.CpuShare, test.CpuShare)
	}
	if ans.Memory != test.Memory {
		return fmt.Sprintf("Memory not right, expect %d m but got %d m", ans.Memory, test.Memory)
	}
	for k, v := range ans.Port {
		if test.Port[k] != v {
			return fmt.Sprintf("Port config not right, expect %s:%s but got %s", k, v, test.Port[k])
		}
	}
	for k, v := range ans.Volume {
		if test.Volume[k] != v {
			return fmt.Sprintf("Volume config not right, expect '%s':'%s' but got '%s'", k, v, test.Volume[k])
		}
	}
	if ans.Images != "" && test.Images != ans.Images {
		return fmt.Sprintf("Images not right, expect '%s' but got '%s'.", ans.Images, test.Images)
	}
	if ans.Command != "" && test.Command != ans.Command {
		return fmt.Sprintf("Command not right, expect '%s' but got '%s'.", ans.Command, test.Command)
	}
	if ans.Command == "" && test.Command != "" {
		return fmt.Sprintf("Unexpect command: %s", test.Command)
	}
	if len(ans.Arg) != len(test.Arg) {
		return fmt.Sprintf("Arguments number not right, expect %d but got %d", len(ans.Arg), len(test.Arg))
	}
	for i := 0; i < len(ans.Arg); i++ { //it operation must place after compare length
		if ans.Arg[i] != test.Arg[i] {
			return fmt.Sprintf("Arguments not right, expect '%s' but got '%s'.", ans.Arg[i], test.Arg[i])
		}
	}
	return ""
}

//===================================================================

//Process the docker command from a string into an array
func splitCommand(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	cmd = strings.TrimPrefix(cmd, "sudo")
	cmd = strings.TrimSpace(cmd)
	newLineReg, _ := regexp.Compile(` \\\n`)
	cmd = newLineReg.ReplaceAllLiteralString(cmd, " ")
	blankReg, _ := regexp.Compile(`[ ]+`)
	cmdArray := blankReg.Split(cmd, -1)
	return cmdArray
}

//check if the name of images is legal
//Rule: repository name must be lowercase letter or number and '_', tag of images can be number or letter and '_','.'
func isImagesName(name string) bool {
	legalReg, _ := regexp.Compile(`^[a-z_0-9\/]+[:]?[\w]+$`)
	return legalReg.MatchString(name)
}

//judge if a port argument is legal, such as 8080:3434 is right
func isPortArg(arg string) bool {
	legalReg, _ := regexp.Compile(`^[\d]{1,5}:[\d]{1,5}$`)
	return legalReg.MatchString(arg)
}

//judge whether a arguments can be used by flag -a or --attach
func isAttach(arg string) bool {
	arg = strings.ToLower(arg)
	if arg == "stdin" || arg == "stderr" || arg == "stdout" {
		return true
	}
	return false
}

//judge if a path argument can be userd by flag -v or --volume
func isDirPath(path string) bool {
	legalReg, _ := regexp.Compile(`^[^: ]+:[^: ]+$`)
	return legalReg.MatchString(path)
}

//judge if the name of container is legal, only [a-zA-Z0-9][a-zA-Z0-9_.-] are allowed.
func isContainerName(name string) bool {
	legalReg, _ := regexp.Compile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)
	return legalReg.MatchString(name)
}

//check if the argument of -workplace flag
func isWorkDir(dir string) bool {
	legalReg, _ := regexp.Compile(`^[\w\/\=\+\-\*\!\\~.#@]+$`)
	return legalReg.MatchString(dir)
}

//check if the argument can be used by flag -m or --menory
func isMemory(arg string) bool {
	reg, _ := regexp.Compile(`^\d{1,30}[bBkKmMgG]{1}$`)
	return reg.MatchString(arg)
}

//return a string array have contain a specified string
func findInArray(array []string, target string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == target {
			return true
		}
	}
	return false
}

//change the string from "abc" to abc
func trimStr(str string) string {
	Reg, _ := regexp.Compile(`^"[^"]*"$`)
	if Reg.MatchString(str) {
		return strings.Trim(str, "\"")
	}
	Reg, _ = regexp.Compile(`^'[^']*'$`)
	if Reg.MatchString(str) {
		return strings.Trim(str, "'")
	}
	return str
}
