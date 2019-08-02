package DockerRun

import (
	"testing"
)

//the example that with uncommon synatax but all have to pass NewContainer()
var passExample = []string{
	`sudo docker run --rm=true -i=true -d=false alpine`,
	`sudo docker run --name="hello" alpine`,
	`sudo docker run --detach=false alpine`,
	`sudo docker run  -itv /dddsss:/workplace   alpine:latest sh -c 'pwd'`,
	`sudo docker run -ip 8080:8080 --rm -dp 8383:8333 alpine:latest`,
	`docker run -itdp8080:7373  alpine:latest`,
	`docker run --rm -it -p8080:7373 -p8081:7373  alpine:latest`,
	`sudo docker run --rm -it -hubuntu  alpine:latest`,
	`sudo docker run --rm -itPp8080:8080  alpine:latest`,
	`docker run --detach -v=$PWD/data:/var/lib/postgresql/data kthiinqwrs/1000010021_postgres`,
}

//the docker command use in the program that as a answer template, all much have to pass NewContainer()
var usingExample = []string{
	`docker run --rm -v $PWD:/workspace username/1000010021_angular ng build`,
	`docker run -d -v $PWD/data:/var/lib/postgresql/data username/1000010021_postgres`,
	`docker run -d -m 1024m -p 8081:8080 username/1000010021_server`,
	`docker run -d --name database -v username_vol:/var/lib/postgresql/data username/1000010022_postgres:latest`,
	`docker run -d --network netname --name dockername username/1000010023_postgres:latest `,
	`docker run -v $PWD/data:/data --name server1 username/1000010024_server:latest`,
	`docker run -t --rm -w /data -v username_vol:/data username/1000010024_server:latest touch hello.txt`,
	`docker run -d -w /data --name server3 -v username_vol:/data username/1000010024_server:latest rm hello.txt`,
}

//the fail example, all of them should unpass
var failExample = []string{
	`sudo docker run  -ipt 8080:8080 alpine:latest`,                          //invalid publish opts format (should be port1:port2 but got 't').
	`docker run -ip 8080:8080 --rmv /work:/work -tp 8383:8333 alpine:latest`, //Unknown flag: --rmv
	`docker run -ip 8080:8080 --rm -v -tp 8383:8333 alpine:latest`,           //Invalid volume argument: -tp
	`sudo docker run`,    //Can't find images name from given command!
	`docker run alpine:`, //Images name alpine: not legal!
	`docker run --rm -it -p8080:7373 -p8080:2344 alpine:latest`, // Port is already allocated: 8080
	` docker run --rm -it--hostname ubuntu  alpine:latest`,      //unknown shorthand flag: - in -hostname
	`sudo docker run --rm-it  alpine:latest`,                    //Unknown flag: --rm-it
	`sudo docker run --rm -it-memory 5MB  alpine:latest`,        //unknown shorthand flag: - in memory
	`docker run -p ="/hello:/world alpine"`,
}

//testint whether NewContainer() can let all correct command pass and return error in all  worng demo
func TestNewContainer(t *testing.T) {
	for _, cmd := range passExample {
		_, err := NewMockContainer(cmd)
		if err != nil {
			t.Fatalf("right command '%s' unpass! reason: %v", cmd, err)
		}
	}
	for _, cmd := range failExample {
		_, err := NewMockContainer(cmd)
		if err == nil {
			t.Fatalf("worng command %s pass! reason: %v", cmd, err)
		}
	}
	for _, cmd := range usingExample {
		_, err := NewMockContainer(cmd)
		if err != nil {
			t.Fatalf("template command %s unpass! reason: %v", cmd, err)
		}
	}
}

func TestChecking1(t *testing.T) {
	standard := `docker run --rm -it --name testcase -v /hello:/world images:latest`
	challenger := []string{
		`docker run -i=true -t=true --rm=true --name="testcase" -v=/hello:/world images:latest`,
		`sudo docker run -itv/hello:/world --rm --name=testcase images`,
		`docker run --interactive --detach --rm --name=testcase --tty -v /hello:/world images`,
		`docker run --rm -it --name testcase -v /hello:/world images:latest`,
		`docker run --interactive=true --detach=false --rm=true --name=testcase --tty -v /hello:/world images`,
		`docker run -i=true -t=true --rm=true --name="testcase" -v=/hello:/world images:latest`,
		`docker run --rm -it --name='testcase' -v /hello:/world images:latest`,
		`docker run --rm -it --name 'testcase' -v /hello:/world images:latest`,
		`docker run --rm -it --name 'testcase' -v=/hello:/world images:latest`,
		`docker run --rm -it --name 'testcase' -v="/hello:/world" images:latest`,
		`docker run --rm -it --name 'testcase' -v "/hello:/world" images:latest`,
	}
	stdcon, err := NewMockContainer(standard)
	if err != nil {
		t.Fatalf("Standard command fall: %v", err)
	}
	for i := 0; i < len(challenger); i++ {
		tempCon, err := NewMockContainer(challenger[i])
		if err != nil {
			t.Fatalf("Create container fail at command %d  : %v", i, err)
		}
		res := Judge(&tempCon, &stdcon)
		if res != "" {
			tempCon.Printf()
			t.Fatalf("Unpass at command %d : %v", i, res)
		}
	}
}

func TestChecking2(t *testing.T) {
	standard := `docker run -p 8080:8080 -w /hello -m 1024m -c 20 images:latest`
	challenger := []string{
		`sudo docker run -w/hello -p8080:8080 -m1024m -c20 images`,
		`sudo docker run -c20  -p8080:8080 -w /hello -m 1g  images:latest`,
		`sudo docker run -w=/hello -p8080:8080 -m1024m -c20 images`,
		`sudo docker run -w="/hello" -p8080:8080 -m1024m -c20 images`,
		`sudo docker run -w "/hello" -p=8080:8080 -m=1024m -c=20 images`,
		`sudo docker run -w "/hello" -p="8080:8080" -m=1024m -c=20 images`,
	}

	stdcon, err := NewMockContainer(standard)
	if err != nil {
		t.Fatalf("Standard command fall: %v", err)
	}
	for i := 0; i < len(challenger); i++ {
		tempCon, err := NewMockContainer(challenger[i])
		if err != nil {
			t.Fatalf("Create container fail at %s : %v", challenger[i], err)
		}
		res := Judge(&tempCon, &stdcon)
		if res != "" {
			t.Fatalf("Unpass at %d command %s : %v", i, challenger[i], res)
		}
	}
}
