//Package controlDriver issues commands based on controller input
package controlDriver

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/AscendTech4H/AscendTechROV/go/commander"
	"github.com/AscendTech4H/AscendTechROV/go/communication"
	"github.com/AscendTech4H/AscendTechROV/go/controller"
	"github.com/AscendTech4H/AscendTechROV/go/debug"
	"github.com/AscendTech4H/AscendTechROV/go/motor"
	"github.com/AscendTech4H/AscendTechROV/go/motor/cmdmotor"
	"github.com/AscendTech4H/AscendTechROV/go/startup"
	"github.com/AscendTech4H/AscendTechROV/go/util"
)

//Add more motors when I know which they are
var robot struct {
	left, right       motor.Motor
	topfront, topback motor.Motor
	claw              struct {
		roll motor.Motor
		grab motor.Motor
	}
}

//Motor IDs
//Note: put them in index order for iota to assign indexes (stating with 0)
const (
	motortf = iota
	motortb
	motorl
	motorr
	motorroll
	motorgrab
)

func init() {
	startup.NewTask(150, func() error {
		robot.left = cmdmotor.Motor(communication.Sender, motorl, motor.DC)
		robot.right = cmdmotor.Motor(communication.Sender, motorr, motor.DC)

		robot.topfront = cmdmotor.Motor(communication.Sender, motortf, motor.DC)
		robot.topback = cmdmotor.Motor(communication.Sender, motortb, motor.DC)

		robot.claw.roll = cmdmotor.Motor(communication.Sender, motorroll, motor.DC)
		robot.claw.grab = cmdmotor.Motor(communication.Sender, motorgrab, motor.Servo)
		return nil
	})
	startup.NewTask(253, func() error {
		f, err := os.Create("debug.log")
		util.UhOh(err)
		q := log.New(f, "meme", 0)
		if communication.Sender != nil {
			tick := time.NewTicker(time.Second / 5)
			go func() {
				for range tick.C {
					debug.VLog("Motor update")
					rob := controller.RobotState()
					l, r := motorCalcFwd(rob.Forward, rob.Turn)
					log.Println(l, r)
					//a := uint8(rangeMap(r, -127, 127, 0, 255))
					//b := uint8(rangeMap(l, -127, 127, 0, 255))
					a := uint8(r)
					b := uint8(l)
					robot.right.Set(a)
					robot.left.Set(b)
					if rob.Tilt == 0 {
						u := uint8(rangeMap(rob.Up, -50, 50, 0, 255))
						robot.topback.Set(u)
						robot.topfront.Set(u)
					} else {
						m := rangeMap(rob.Tilt, -90, 90, -255, 255)
						a := m
						if a < 0 {
							a *= -1
						}
						var f, b uint8 = uint8(a), uint8(a)
						if m > 0 {
							b = 255 - b
						} else {
							f = 255 - f
						}
						robot.topfront.Set(f)
						robot.topback.Set(b)
					}
					c := uint8(0)
					switch rob.ClawTurn {
					case controller.CCW:
						c = 127 - 50
					case controller.CW:
						c = 127 + 50
					case controller.STOP:
						c = 127
					}
					robot.claw.roll.Set(c)
					if rob.Claw {
						robot.claw.grab.Set(90)
					} else {
						robot.claw.grab.Set(180)
					}
					communication.Bus.AsSender().Send(commander.SetLaser(rob.Laser))
					q.Println([]uint8{robot.left.State(), robot.right.State(), robot.topback.State(), robot.topfront.State(), robot.claw.grab.State(), robot.claw.roll.State()})
				}
			}()
		}
		return nil
	})
}

func rangeMap(in, inmin, inmax, outmin, outmax int) int {
	return (((in - inmin) * (outmax - outmin)) / (inmax - inmin)) + outmin
}

func motorCalcFwd(X, Y int) (l, r int) {
	x := float64(X)
	y := float64(Y)
	ang := math.Atan2(y, x) - math.Pi/4
	mag := math.Hypot(x/50, y/50)
	r = int(mag*math.Cos(ang)*127.5 + 127.5)
	if r == 256 {
		r = r - 1
	}
	ang = math.Atan2(y, x) - 3*math.Pi/4
	l = int(mag*math.Cos(ang)*127.5 + 127.5)
	if l == 256 {
		l = l - 1
	}
	return
}
