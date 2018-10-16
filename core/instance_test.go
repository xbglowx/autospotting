package autospotting

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/davecgh/go-spew/spew"
)

func TestMake(t *testing.T) {
	expected := instanceMap{}
	is := &instanceManager{}

	is.make()
	if !reflect.DeepEqual(is.catalog, expected) {
		t.Errorf("Catalog's type: '%s' expected: '%s'",
			reflect.TypeOf(is.catalog).String(),
			reflect.TypeOf(expected).String())
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		catalog  instanceMap
		expected instanceMap
	}{
		{name: "map contains a nil pointer",
			catalog: instanceMap{
				"inst1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
				"inst2": nil,
			},
			expected: instanceMap{
				"1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
			},
		},
		{name: "map has 1 instance",
			catalog: instanceMap{
				"inst1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
			},
			expected: instanceMap{
				"1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
			},
		},
		{name: "map has several instances",
			catalog: instanceMap{
				"inst1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
				"inst2": {Instance: &ec2.Instance{InstanceId: aws.String("2")}},
				"inst3": {Instance: &ec2.Instance{InstanceId: aws.String("3")}},
			},
			expected: instanceMap{
				"1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
				"2": {Instance: &ec2.Instance{InstanceId: aws.String("2")}},
				"3": {Instance: &ec2.Instance{InstanceId: aws.String("3")}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &instanceManager{}
			is.make()
			for _, c := range tt.catalog {
				is.add(c)
			}
			if !reflect.DeepEqual(tt.expected, is.catalog) {
				t.Errorf("Value received: %v expected %v", is.catalog, tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		catalog  instanceMap
		idToGet  string
		expected *instance
	}{
		{name: "map contains the required instance",
			catalog: instanceMap{
				"inst1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
				"inst2": {Instance: &ec2.Instance{InstanceId: aws.String("2")}},
				"inst3": {Instance: &ec2.Instance{InstanceId: aws.String("3")}},
			},
			idToGet:  "inst2",
			expected: &instance{Instance: &ec2.Instance{InstanceId: aws.String("2")}},
		},
		{name: "catalog doesn't contain the instance",
			catalog: instanceMap{
				"inst1": {Instance: &ec2.Instance{InstanceId: aws.String("1")}},
				"inst2": {Instance: &ec2.Instance{InstanceId: aws.String("2")}},
				"inst3": {Instance: &ec2.Instance{InstanceId: aws.String("3")}},
			},
			idToGet:  "7",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &instanceManager{}
			is.make()
			is.catalog = tt.catalog
			retInstance := is.get(tt.idToGet)
			if !reflect.DeepEqual(tt.expected, retInstance) {
				t.Errorf("Value received: %v expected %v", retInstance, tt.expected)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name     string
		catalog  instanceMap
		expected int
	}{
		{name: "map is nil",
			catalog:  nil,
			expected: 0,
		},
		{name: "map is empty",
			catalog:  instanceMap{},
			expected: 0,
		},
		{name: "map has 1 instance",
			catalog: instanceMap{
				"id-1": {},
			},
			expected: 1,
		},
		{name: "map has several instances",
			catalog: instanceMap{
				"id-1": {},
				"id-2": {},
				"id-3": {},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &instanceManager{}
			is.catalog = tt.catalog
			ret := is.count()
			if ret != tt.expected {
				t.Errorf("Value received: '%d' expected %d", ret, tt.expected)
			}
		})
	}
}

func TestCount64(t *testing.T) {
	tests := []struct {
		name     string
		catalog  instanceMap
		expected int64
	}{
		{name: "map is nil",
			catalog:  nil,
			expected: 0,
		},
		{name: "map is empty",
			catalog:  instanceMap{},
			expected: 0,
		},
		{name: "map has 1 instance",
			catalog: instanceMap{
				"id-1": {},
			},
			expected: 1,
		},
		{name: "map has several instances",
			catalog: instanceMap{
				"id-1": {},
				"id-2": {},
				"id-3": {},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &instanceManager{}
			is.catalog = tt.catalog
			ret := is.count64()
			if ret != tt.expected {
				t.Errorf("Value received: '%d' expected %d", ret, tt.expected)
			}
		})
	}
}

func TestIsSpot(t *testing.T) {

	tests := []struct {
		name      string
		lifeCycle *string
		expected  bool
	}{
		{name: "LifeCycle is nil",
			lifeCycle: nil,
			expected:  false,
		},
		{name: "LifeCycle is not nil but not spot",
			lifeCycle: aws.String("something"),
			expected:  false,
		},
		{name: "LifeCycle is not nil and is spot",
			lifeCycle: aws.String("spot"),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &instance{Instance: &ec2.Instance{}}
			i.InstanceLifecycle = tt.lifeCycle
			retValue := i.isSpot()
			if retValue != tt.expected {
				if tt.lifeCycle != nil {
					t.Errorf("Value received for '%v': %t expected %t", *tt.lifeCycle, retValue, tt.expected)
				} else {
					t.Errorf("Value received for '%v': %t expected %t", tt.lifeCycle, retValue, tt.expected)
				}
			}
		})
	}
}
func TestIsEBSCompatible(t *testing.T) {
	tests := []struct {
		name         string
		spotInfo     instanceTypeInformation
		instanceInfo instance
		expected     bool
	}{
		{name: "EBS not Optimized Spot not Optimized",
			spotInfo: instanceTypeInformation{
				hasEBSOptimization: false,
			},
			instanceInfo: instance{
				Instance: &ec2.Instance{
					EbsOptimized: nil,
				},
			},
			expected: true,
		},
		{name: "EBS Optimized Spot Optimized",
			spotInfo: instanceTypeInformation{
				hasEBSOptimization: true,
			},
			instanceInfo: instance{
				Instance: &ec2.Instance{
					EbsOptimized: &[]bool{true}[0],
				},
			},
			expected: true,
		},
		{name: "EBS Optimized Spot not Optimized",
			spotInfo: instanceTypeInformation{
				hasEBSOptimization: false,
			},
			instanceInfo: instance{
				Instance: &ec2.Instance{
					EbsOptimized: &[]bool{true}[0],
				},
			},
			expected: false,
		},
		{name: "EBS not Optimized Spot Optimized",
			spotInfo: instanceTypeInformation{
				hasEBSOptimization: true,
			},
			instanceInfo: instance{
				Instance: &ec2.Instance{
					EbsOptimized: &[]bool{false}[0],
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &tt.instanceInfo
			retValue := i.isEBSCompatible(tt.spotInfo)
			if retValue != tt.expected {
				t.Errorf("Value received: %t expected %t", retValue, tt.expected)
			}
		})
	}
}

func TestIsPriceCompatible(t *testing.T) {
	tests := []struct {
		name             string
		spotPrices       prices
		availabilityZone *string
		instancePrice    float64
		bestPrice        float64
		expected         bool
	}{
		{name: "No spot price for such availability zone",
			spotPrices: prices{
				spot: map[string]float64{
					"eu-central-1": 0.5,
					"eu-west-1":    1.0,
					"eu-west-2":    2.0,
				},
			},
			availabilityZone: aws.String("eu-west-42"),
			instancePrice:    5.0,
			bestPrice:        0.7,
			expected:         false,
		},
		{name: "Spot price is higher than bestPrice",
			spotPrices: prices{
				spot: map[string]float64{
					"eu-central-1": 0.5,
					"eu-west-1":    1.0,
					"eu-west-2":    2.0,
				},
			},
			availabilityZone: aws.String("eu-west-1"),
			instancePrice:    5.0,
			bestPrice:        0.7,
			expected:         false,
		},
		{name: "Spot price is lower than bestPrice",
			spotPrices: prices{
				spot: map[string]float64{
					"eu-central-1": 0.5,
					"eu-west-1":    1.0,
					"eu-west-2":    2.0,
				},
			},
			availabilityZone: aws.String("eu-west-1"),
			instancePrice:    5.0,
			bestPrice:        1.4,
			expected:         true,
		},
		{name: "Spot price is 0.0",
			spotPrices: prices{
				spot: map[string]float64{
					"eu-central-1": 0.0,
					"eu-west-1":    0.0,
					"eu-west-2":    0.0,
				},
			},
			availabilityZone: aws.String("eu-west-1"),
			instancePrice:    5.0,
			bestPrice:        1.4,
			expected:         false,
		},
		{name: "Spot price is higher than instance price",
			spotPrices: prices{
				spot: map[string]float64{
					"eu-central-1": 0.5,
					"eu-west-1":    1.0,
					"eu-west-2":    2.0,
				},
			},
			availabilityZone: aws.String("eu-west-1"),
			instancePrice:    0.7,
			bestPrice:        0.7,
			expected:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &instance{Instance: &ec2.Instance{
				Placement: &ec2.Placement{
					AvailabilityZone: tt.availabilityZone,
				}},
				price: tt.instancePrice,
			}
			candidate := instanceTypeInformation{pricing: prices{}}
			candidate.pricing = tt.spotPrices
			spotPrice := i.calculatePrice(candidate)
			retValue := i.isPriceCompatible(spotPrice, tt.bestPrice)
			if retValue != tt.expected {
				t.Errorf("Value received: %t expected %t", retValue, tt.expected)
			}
		})
	}
}

func TestIsClassCompatible(t *testing.T) {
	tests := []struct {
		name           string
		spotInfo       instanceTypeInformation
		instanceCPU    int
		instanceMemory float32
		instanceGPU    int
		expected       bool
	}{
		{name: "Spot is higher in both CPU & memory",
			spotInfo: instanceTypeInformation{
				vCPU:   10,
				memory: 2.5,
			},
			instanceCPU:    5,
			instanceMemory: 1.0,
			expected:       true,
		},
		{name: "Spot is lower in CPU but higher in memory",
			spotInfo: instanceTypeInformation{
				vCPU:   10,
				memory: 2.5,
			},
			instanceCPU:    15,
			instanceMemory: 1.0,
			expected:       false,
		},
		{name: "Spot is lower in memory but higher in CPU",
			spotInfo: instanceTypeInformation{
				vCPU:   10,
				memory: 2.5,
			},
			instanceCPU:    5,
			instanceMemory: 10.0,
			expected:       false,
		},
		{name: "Spot is lower in both CPU & memory",
			spotInfo: instanceTypeInformation{
				vCPU:   10,
				memory: 2.5,
			},
			instanceCPU:    15,
			instanceMemory: 5.0,
			expected:       false,
		},
		{name: "Spot is lower in CPU, memory and GPU ",
			spotInfo: instanceTypeInformation{
				vCPU:   10,
				memory: 2.5,
				GPU:    0,
			},
			instanceCPU:    15,
			instanceMemory: 5.0,
			instanceGPU:    1,
			expected:       false,
		},

		{name: "Spot is higher in CPU, memory and GPU ",
			spotInfo: instanceTypeInformation{
				vCPU:   10,
				memory: 20,
				GPU:    4,
			},
			instanceCPU:    8,
			instanceMemory: 4,
			instanceGPU:    2,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &instance{typeInfo: instanceTypeInformation{
				vCPU:   tt.instanceCPU,
				memory: tt.instanceMemory,
			},
			}
			retValue := i.isClassCompatible(tt.spotInfo)
			if retValue != tt.expected {
				t.Errorf("Value received: %t expected %t", retValue, tt.expected)
			}
		})
	}
}

func TestIsStorageCompatible(t *testing.T) {
	tests := []struct {
		name            string
		spotInfo        instanceTypeInformation
		instanceInfo    instanceTypeInformation
		attachedVolumes int
		expected        bool
	}{
		{name: "Instance has no attached volumes",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 0.0,
				instanceStoreDeviceSize:  0.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 0.0,
				instanceStoreIsSSD:      false,
			},
			attachedVolumes: 0,
			expected:        true,
		},
		{name: "Spot's storage is identical to instance",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 1,
				instanceStoreDeviceSize:  50.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      false,
			},
			attachedVolumes: 1,
			expected:        true,
		},
		{name: "Spot's storage is lower than the instance's one",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 1,
				instanceStoreDeviceSize:  25.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      false,
			},
			attachedVolumes: 1,
			expected:        false,
		},
		{name: "Spot's storage is bigger than the instance's one",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 1,
				instanceStoreDeviceSize:  150.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      false,
			},
			attachedVolumes: 1,
			expected:        true,
		},
		{name: "Spot's storage is bigger and has less storage attached",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 1,
				instanceStoreDeviceSize:  150.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      false,
			},
			attachedVolumes: 2,
			expected:        false,
		},
		{name: "Spot's storage is bigger and has more storage attached",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 4,
				instanceStoreDeviceSize:  150.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      false,
			},
			attachedVolumes: 1,
			expected:        true,
		},
		{name: "Spot's storage is bigger and has more storage attached but isn't SSD (unlike the instance)",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 4,
				instanceStoreDeviceSize:  150.0,
				instanceStoreIsSSD:       false,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      true,
			},
			attachedVolumes: 1,
			expected:        false,
		},
		{name: "Spot's storage is bigger, has more storage attached, is SSD (like the instance)",
			spotInfo: instanceTypeInformation{
				instanceStoreDeviceCount: 4,
				instanceStoreDeviceSize:  150.0,
				instanceStoreIsSSD:       true,
			},
			instanceInfo: instanceTypeInformation{
				instanceStoreDeviceSize: 50.0,
				instanceStoreIsSSD:      true,
			},
			attachedVolumes: 1,
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &instance{typeInfo: tt.instanceInfo}
			retValue := i.isStorageCompatible(tt.spotInfo, tt.attachedVolumes)
			if retValue != tt.expected {
				t.Errorf("Value received: %t expected %t", retValue, tt.expected)
			}
		})
	}
}

func TestIsVirtualizationCompatible(t *testing.T) {
	tests := []struct {
		name                       string
		spotVirtualizationTypes    []string
		instanceVirtualizationType *string
		expected                   bool
	}{
		{name: "Spot's virtualization includes instance's one (pv case)",
			spotVirtualizationTypes:    []string{"PV", "HVM"},
			instanceVirtualizationType: aws.String("paravirtual"),
			expected:                   true,
		},
		{name: "Spot's virtualization includes instance's one (hvm case)",
			spotVirtualizationTypes:    []string{"PV", "HVM"},
			instanceVirtualizationType: aws.String("hvm"),
			expected:                   true,
		},
		{name: "Spot's virtualization does not include any type",
			spotVirtualizationTypes:    []string{},
			instanceVirtualizationType: aws.String("hvm"),
			expected:                   true,
		},
		{name: "Spot's virtualization doesn't include instance's one (pv case)",
			spotVirtualizationTypes:    []string{"HVM"},
			instanceVirtualizationType: aws.String("paravirtual"),
			expected:                   false,
		},
		{name: "Spot's virtualization doesn't include instance's one (hvm case)",
			spotVirtualizationTypes:    []string{"PV"},
			instanceVirtualizationType: aws.String("hvm"),
			expected:                   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &instance{Instance: &ec2.Instance{
				VirtualizationType: tt.instanceVirtualizationType,
			}}
			retValue := i.isVirtualizationCompatible(tt.spotVirtualizationTypes)
			if retValue != tt.expected {
				t.Errorf("Value received: %t expected %t", retValue, tt.expected)
			}
		})
	}
}

func TestGetCheapestCompatibleSpotInstanceType(t *testing.T) {
	tests := []struct {
		name           string
		spotInfos      map[string]instanceTypeInformation
		instanceInfo   *instance
		asg            *autoScalingGroup
		expectedString string
		expectedError  error
		allowedList    []string
		disallowedList []string
	}{
		{name: "better/cheaper spot instance found",
			spotInfos: map[string]instanceTypeInformation{
				"1": {
					instanceType: "type1",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.5,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
				"2": {
					instanceType: "type2",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.8,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
			},
			instanceInfo: &instance{
				Instance: &ec2.Instance{
					VirtualizationType: aws.String("paravirtual"),
					Placement: &ec2.Placement{
						AvailabilityZone: aws.String("eu-central-1"),
					},
				},
				typeInfo: instanceTypeInformation{
					instanceType:             "typeX",
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
				},
				price:  0.75,
				region: &region{},
			},
			asg: &autoScalingGroup{
				name: "test-asg",
				instances: makeInstancesWithCatalog(
					instanceMap{
						"id-1": {
							Instance: &ec2.Instance{
								InstanceId:        aws.String("id-1"),
								InstanceType:      aws.String("typeX"),
								Placement:         &ec2.Placement{AvailabilityZone: aws.String("eu-west-1")},
								InstanceLifecycle: aws.String("spot"),
							},
						},
					},
				),
				Group: &autoscaling.Group{
					DesiredCapacity: aws.Int64(4),
				},
			},
			expectedString: "type1",
			expectedError:  nil,
		},
		{name: "better/cheaper spot instance found but marked as disallowed",
			spotInfos: map[string]instanceTypeInformation{
				"1": {
					instanceType: "type1",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.5,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
				"2": {
					instanceType: "type2",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.8,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
			},
			instanceInfo: &instance{
				Instance: &ec2.Instance{
					VirtualizationType: aws.String("paravirtual"),
					Placement: &ec2.Placement{
						AvailabilityZone: aws.String("eu-central-1"),
					},
				},
				typeInfo: instanceTypeInformation{
					instanceType:             "typeX",
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
				},
				price:  0.75,
				region: &region{},
			},
			asg: &autoScalingGroup{
				name: "test-asg",
				instances: makeInstancesWithCatalog(
					instanceMap{
						"id-1": {
							Instance: &ec2.Instance{
								InstanceId:        aws.String("id-1"),
								InstanceType:      aws.String("typeX"),
								Placement:         &ec2.Placement{AvailabilityZone: aws.String("eu-west-1")},
								InstanceLifecycle: aws.String("spot"),
							},
						},
					},
				),
				Group: &autoscaling.Group{
					DesiredCapacity: aws.Int64(4),
				},
			},
			disallowedList: []string{"type*"},
			expectedString: "",
			expectedError:  errors.New("No cheaper spot instance types could be found"),
		},
		{name: "better/cheaper spot instance found but not marked as allowed",
			spotInfos: map[string]instanceTypeInformation{
				"1": {
					instanceType: "type1",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.5,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
				"2": {
					instanceType: "type2",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.8,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
			},
			instanceInfo: &instance{
				Instance: &ec2.Instance{
					VirtualizationType: aws.String("paravirtual"),
					Placement: &ec2.Placement{
						AvailabilityZone: aws.String("eu-central-1"),
					},
				},
				typeInfo: instanceTypeInformation{
					instanceType:             "typeX",
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
				},
				price:  0.75,
				region: &region{},
			},
			asg: &autoScalingGroup{
				name: "test-asg",
				instances: makeInstancesWithCatalog(
					instanceMap{
						"id-1": {
							Instance: &ec2.Instance{
								InstanceId:        aws.String("id-1"),
								InstanceType:      aws.String("typeX"),
								Placement:         &ec2.Placement{AvailabilityZone: aws.String("eu-west-1")},
								InstanceLifecycle: aws.String("spot"),
							},
						},
					},
				),
				Group: &autoscaling.Group{
					DesiredCapacity: aws.Int64(4),
				},
			},
			allowedList:    []string{"asdf*"},
			expectedString: "",
			expectedError:  errors.New("No cheaper spot instance types could be found"),
		},
		{name: "better/cheaper spot instance found and marked as allowed",
			spotInfos: map[string]instanceTypeInformation{
				"1": {
					instanceType: "type1",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.5,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
				"2": {
					instanceType: "type2",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.8,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
			},
			instanceInfo: &instance{
				Instance: &ec2.Instance{
					VirtualizationType: aws.String("paravirtual"),
					Placement: &ec2.Placement{
						AvailabilityZone: aws.String("eu-central-1"),
					},
				},
				typeInfo: instanceTypeInformation{
					instanceType:             "typeX",
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
				},
				price:  0.75,
				region: &region{},
			},
			asg: &autoScalingGroup{
				name: "test-asg",
				instances: makeInstancesWithCatalog(
					instanceMap{
						"id-1": {
							Instance: &ec2.Instance{
								InstanceId:        aws.String("id-1"),
								InstanceType:      aws.String("typeX"),
								Placement:         &ec2.Placement{AvailabilityZone: aws.String("eu-west-1")},
								InstanceLifecycle: aws.String("spot"),
							},
						},
					},
				),
				Group: &autoscaling.Group{
					DesiredCapacity: aws.Int64(4),
				},
			},

			allowedList:    []string{"ty*"},
			expectedString: "type1",
			expectedError:  nil,
		},
		{name: "better/cheaper spot instance not found",
			spotInfos: map[string]instanceTypeInformation{
				"1": {
					instanceType: "type1",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.5,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
				"2": {
					instanceType: "type2",
					pricing: prices{
						spot: map[string]float64{
							"eu-central-1": 0.8,
							"eu-west-1":    1.0,
							"eu-west-2":    2.0,
						},
					},
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
					virtualizationTypes:      []string{"PV", "else"},
				},
			},
			instanceInfo: &instance{
				Instance: &ec2.Instance{
					VirtualizationType: aws.String("paravirtual"),
					Placement: &ec2.Placement{
						AvailabilityZone: aws.String("eu-central-1"),
					},
				},
				typeInfo: instanceTypeInformation{
					instanceType:             "typeX",
					vCPU:                     10,
					memory:                   2.5,
					instanceStoreDeviceCount: 1,
					instanceStoreDeviceSize:  50.0,
					instanceStoreIsSSD:       false,
				},
				price:  0.45,
				region: &region{},
			},
			asg: &autoScalingGroup{
				name: "test-asg",
				instances: makeInstancesWithCatalog(
					instanceMap{
						"id-1": {
							Instance: &ec2.Instance{
								InstanceId:        aws.String("id-1"),
								InstanceType:      aws.String("typeX"),
								Placement:         &ec2.Placement{AvailabilityZone: aws.String("eu-west-1")},
								InstanceLifecycle: aws.String("spot"),
							},
						},
					},
				),
				Group: &autoscaling.Group{
					DesiredCapacity: aws.Int64(4),
				},
			},
			expectedString: "",
			expectedError:  errors.New("No cheaper spot instance types could be found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := tt.instanceInfo
			i.region.instanceTypeInformation = tt.spotInfos
			i.asg = tt.asg
			allowedList := tt.allowedList
			disallowedList := tt.disallowedList
			retValue, err := i.getCheapestCompatibleSpotInstanceType(allowedList, disallowedList)
			if err == nil && tt.expectedError != err {
				t.Errorf("Error received: %v expected %v", err, tt.expectedError.Error())
			} else if err != nil && tt.expectedError == nil {
				t.Errorf("Error received: %s expected %s", err.Error(), tt.expectedError)
			} else if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("Error received: %s expected %s", err.Error(), tt.expectedError.Error())
			} else if retValue.instanceType != tt.expectedString {
				t.Errorf("Value received: %s expected %s", retValue.instanceType, tt.expectedString)
			}
		})
	}
}

func TestGetPricetoBid(t *testing.T) {
	tests := []struct {
		spotPercentage       float64
		currentSpotPrice     float64
		currentOnDemandPrice float64
		policy               string
		want                 float64
	}{
		{
			spotPercentage:       50.0,
			currentSpotPrice:     0.0216,
			currentOnDemandPrice: 0.0464,
			policy:               "aggressive",
			want:                 0.0324,
		},
		{
			spotPercentage:       79.0,
			currentSpotPrice:     0.0216,
			currentOnDemandPrice: 0.0464,
			policy:               "aggressive",
			want:                 0.038664,
		},
		{
			spotPercentage:       79.0,
			currentSpotPrice:     0.0216,
			currentOnDemandPrice: 0.0464,
			policy:               "normal",
			want:                 0.0464,
		},
		{
			spotPercentage:       200.0,
			currentSpotPrice:     0.0216,
			currentOnDemandPrice: 0.0464,
			policy:               "aggressive",
			want:                 0.0464,
		},
	}
	for _, tt := range tests {
		cfg := &Config{
			SpotPriceBufferPercentage: tt.spotPercentage,
			BiddingPolicy:             tt.policy,
		}
		i := &instance{
			region: &region{
				name: "us-east-1",
				conf: cfg,
			},
		}

		currentSpotPrice := tt.currentSpotPrice
		currentOnDemandPrice := tt.currentOnDemandPrice
		actualPrice := i.getPricetoBid(currentOnDemandPrice, currentSpotPrice)
		if math.Abs(actualPrice-tt.want) > 0.000001 {
			t.Errorf("percentage = %.2f, policy = %s, expected price = %.5f, want %.5f, currentSpotPrice = %.5f",
				tt.spotPercentage, tt.policy, actualPrice, tt.want, currentSpotPrice)
		}
	}
}

func TestTerminate(t *testing.T) {
	tests := []struct {
		name     string
		tags     []*ec2.Tag
		inst     *instance
		expected error
	}{
		{
			name: "no issue with terminate",
			tags: []*ec2.Tag{},
			inst: &instance{
				Instance: &ec2.Instance{
					InstanceId: aws.String("id1"),
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNameRunning),
					},
				},
				region: &region{
					services: connections{
						ec2: mockEC2{
							tierr: nil,
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "issue with terminate",
			tags: []*ec2.Tag{},
			inst: &instance{
				Instance: &ec2.Instance{
					InstanceId: aws.String("id1"),
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNameRunning),
					},
				},
				region: &region{
					services: connections{
						ec2: mockEC2{
							tierr: errors.New(""),
						},
					},
				},
			},
			expected: errors.New(""),
		},
	}
	for _, tt := range tests {
		ret := tt.inst.terminate()
		if ret != nil && ret.Error() != tt.expected.Error() {
			t.Errorf("error actual: %s, expected: %s", ret.Error(), tt.expected.Error())
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		y        int
		expected int
	}{
		{name: "Testing min between 0 and 0",
			x:        0,
			y:        0,
			expected: 0,
		},
		{name: "Testing min between 0 and 10",
			x:        0,
			y:        10,
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retValue := min(tt.x, tt.y)
			if retValue != tt.expected {
				t.Errorf("Value received: %d expected %d", retValue, tt.expected)
			}
		})
	}
}

func TestGenerateTagList(t *testing.T) {
	tests := []struct {
		name                     string
		ASGName                  string
		ASGLCName                string
    ASGLTName                string
		instanceTags             []*ec2.Tag
		expectedTagSpecification []*ec2.TagSpecification
	}{
		{name: "no tags on original instance launch configuration",
			ASGLCName:    "testLC0",
			ASGName:      "myASG",
			instanceTags: []*ec2.Tag{},
			expectedTagSpecification: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("LaunchConfigurationName"),
							Value: aws.String("testLC0"),
						},
						{
							Key:   aws.String("launched-by-autospotting"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("launched-for-asg"),
							Value: aws.String("myASG"),
						},
					},
				},
			},
		},
    {name: "no tags on original instance launch template",
      ASGLTName:    "testLT0",
      ASGName:      "myASG",
      instanceTags: []*ec2.Tag{},
      expectedTagSpecification: []*ec2.TagSpecification{
        {
          ResourceType: aws.String("instance"),
          Tags: []*ec2.Tag{
            {
              Key:   aws.String("LaunchTemplateName"),
              Value: aws.String("testLT0"),
            },
            {
              Key:   aws.String("launched-by-autospotting"),
              Value: aws.String("true"),
            },
            {
              Key:   aws.String("launched-for-asg"),
              Value: aws.String("myASG"),
            },
          },
        },
      },
    },
		{name: "Multiple tags on original instance",
			ASGLCName: "testLC0",
			ASGName:   "myASG",
			instanceTags: []*ec2.Tag{
				{
					Key:   aws.String("foo"),
					Value: aws.String("bar"),
				},
				{
					Key:   aws.String("baz"),
					Value: aws.String("bazinga"),
				},
			},
			expectedTagSpecification: []*ec2.TagSpecification{
				{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("LaunchConfigurationName"),
							Value: aws.String("testLC0"),
						},
						{
							Key:   aws.String("launched-by-autospotting"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("launched-for-asg"),
							Value: aws.String("myASG"),
						},
						{
							Key:   aws.String("foo"),
							Value: aws.String("bar"),
						},
						{
							Key:   aws.String("baz"),
							Value: aws.String("bazinga"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			i := instance{
				Instance: &ec2.Instance{
					Tags: tt.instanceTags,
				},
				asg: &autoScalingGroup{
					name: tt.ASGName,
					Group: &autoscaling.Group{
						LaunchConfigurationName: aws.String(tt.ASGLCName),
					},
				},
			}

			tags := i.generateTagsList()
			if !reflect.DeepEqual(tags, tt.expectedTagSpecification) {
				t.Errorf("propagatedInstanceTags received: %+v, expected: %+v",
					tags, tt.expectedTagSpecification)
			}
		})
	}
}

func Test_instance_convertBlockDeviceMappings(t *testing.T) {

	tests := []struct {
		name string
		lc   *launchConfiguration
		want []*ec2.BlockDeviceMapping
	}{
		{
			name: "nil launch configuration",
			lc:   nil,
			want: []*ec2.BlockDeviceMapping{},
		}, {
			name: "nil block device mapping",
			lc: &launchConfiguration{
				LaunchConfiguration: &autoscaling.LaunchConfiguration{
					BlockDeviceMappings: nil,
				},
			},
			want: []*ec2.BlockDeviceMapping{},
		},
		{
			name: "instance-store only",
			lc: &launchConfiguration{
				LaunchConfiguration: &autoscaling.LaunchConfiguration{
					BlockDeviceMappings: []*autoscaling.BlockDeviceMapping{
						{
							DeviceName:  aws.String("/dev/ephemeral0"),
							Ebs:         nil,
							NoDevice:    aws.Bool(false),
							VirtualName: aws.String("foo"),
						},
						{
							DeviceName:  aws.String("/dev/ephemeral1"),
							Ebs:         nil,
							NoDevice:    aws.Bool(false),
							VirtualName: aws.String("bar"),
						},
					},
				},
			},
			want: []*ec2.BlockDeviceMapping{
				{
					DeviceName:  aws.String("/dev/ephemeral0"),
					Ebs:         nil,
					NoDevice:    aws.String("false"),
					VirtualName: aws.String("foo"),
				},
				{
					DeviceName:  aws.String("/dev/ephemeral1"),
					Ebs:         nil,
					NoDevice:    aws.String("false"),
					VirtualName: aws.String("bar"),
				},
			},
		},

		{
			name: "instance-store and EBS",
			lc: &launchConfiguration{
				LaunchConfiguration: &autoscaling.LaunchConfiguration{
					BlockDeviceMappings: []*autoscaling.BlockDeviceMapping{
						{
							DeviceName:  aws.String("/dev/ephemeral0"),
							Ebs:         nil,
							NoDevice:    aws.Bool(false),
							VirtualName: aws.String("foo"),
						},
						{
							DeviceName: aws.String("/dev/xvda"),
							Ebs: &autoscaling.Ebs{
								DeleteOnTermination: aws.Bool(false),
								VolumeSize:          aws.Int64(10),
							},
							VirtualName: aws.String("bar"),
						},
						{
							DeviceName: aws.String("/dev/xvdb"),
							Ebs: &autoscaling.Ebs{
								DeleteOnTermination: aws.Bool(true),
								VolumeSize:          aws.Int64(20),
							},
							VirtualName: aws.String("baz"),
						},
					},
				},
			},
			want: []*ec2.BlockDeviceMapping{
				{
					DeviceName:  aws.String("/dev/ephemeral0"),
					Ebs:         nil,
					NoDevice:    aws.String("false"),
					VirtualName: aws.String("foo"),
				},
				{
					DeviceName: aws.String("/dev/xvda"),
					Ebs: &ec2.EbsBlockDevice{
						DeleteOnTermination: aws.Bool(false),
						VolumeSize:          aws.Int64(10),
					},
					VirtualName: aws.String("bar"),
				},
				{
					DeviceName: aws.String("/dev/xvdb"),
					Ebs: &ec2.EbsBlockDevice{
						DeleteOnTermination: aws.Bool(true),
						VolumeSize:          aws.Int64(20),
					},
					VirtualName: aws.String("baz"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &instance{}
			if got := i.convertBlockDeviceMappings(tt.lc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("instance.convertBlockDeviceMappings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_instance_convertSecurityGroups(t *testing.T) {

	tests := []struct {
		name string
		inst instance
		want []*string
	}{
		{
			name: "missing SGs",
			inst: instance{
				Instance: &ec2.Instance{
					SecurityGroups: []*ec2.GroupIdentifier{},
				},
			},
			want: []*string{},
		},
		{
			name: "single SG",
			inst: instance{
				Instance: &ec2.Instance{
					SecurityGroups: []*ec2.GroupIdentifier{{
						GroupId:   aws.String("sg-123"),
						GroupName: aws.String("foo"),
					}},
				},
			},
			want: []*string{aws.String("sg-123")},
		},
		{
			name: "multiple SGs",
			inst: instance{
				Instance: &ec2.Instance{
					SecurityGroups: []*ec2.GroupIdentifier{{
						GroupId:   aws.String("sg-123"),
						GroupName: aws.String("foo"),
					},
						{
							GroupId:   aws.String("sg-456"),
							GroupName: aws.String("bar"),
						},
					},
				},
			},
			want: []*string{aws.String("sg-123"), aws.String("sg-456")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.inst.convertSecurityGroups(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("instance.convertSecurityGroups() = %v, want %v",
					spew.Sdump(got), spew.Sdump(tt.want))
			}
		})
	}
}

func Test_instance_createRunInstancesInput(t *testing.T) {

	type args struct {
		instanceType string
		price        float64
	}
	tests := []struct {
		name string
		inst instance
		args args
		want *ec2.RunInstancesInput
	}{
		{
			name: "create run instances input without launch-configuration",
			inst: instance{
				asg: &autoScalingGroup{
					name: "mygroup",
					Group: &autoscaling.Group{
						LaunchConfigurationName: aws.String("myLC"),
						LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
							LaunchTemplateId:   aws.String("lt-id"),
							LaunchTemplateName: aws.String("lt-name"),
						},
					},
					launchConfiguration: nil,
				},
				Instance: &ec2.Instance{
					EbsOptimized: aws.Bool(true),

					IamInstanceProfile: &ec2.IamInstanceProfile{
						Arn: aws.String("profile-arn"),
					},

					ImageId:      aws.String("ami-123"),
					InstanceType: aws.String("t2.medium"),
					KeyName:      aws.String("mykey"),

					Placement: &ec2.Placement{
						Affinity: aws.String("foo"),
					},

					SecurityGroups: []*ec2.GroupIdentifier{
						{
							GroupName: aws.String("foo"),
							GroupId:   aws.String("sg-123"),
						},
						{
							GroupName: aws.String("bar"),
							GroupId:   aws.String("sg-456"),
						},
					},

					SubnetId: aws.String("subnet-123"),
				},
			}, args: args{
				instanceType: "t2.small",
				price:        1.5,
			},
			want: &ec2.RunInstancesInput{

				EbsOptimized: aws.Bool(true),

				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Arn: aws.String("profile-arn"),
				},

				ImageId: aws.String("ami-123"),

				InstanceMarketOptions: &ec2.InstanceMarketOptionsRequest{
					MarketType: aws.String("spot"),
					SpotOptions: &ec2.SpotMarketOptions{
						MaxPrice: aws.String("1.5"),
					},
				},

				InstanceType: aws.String("t2.small"),
				KeyName:      aws.String("mykey"),

				LaunchTemplate: &ec2.LaunchTemplateSpecification{
					LaunchTemplateId:   aws.String("lt-id"),
					LaunchTemplateName: aws.String("lt-name"),
				},

				MaxCount: aws.Int64(1),
				MinCount: aws.Int64(1),

				Placement: &ec2.Placement{
					Affinity: aws.String("foo"),
				},

				SecurityGroupIds: []*string{
					aws.String("sg-123"),
					aws.String("sg-456"),
				},

				SubnetId: aws.String("subnet-123"),

				TagSpecifications: []*ec2.TagSpecification{{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("LaunchConfigurationName"),
							Value: aws.String("myLC"),
						},
						{
							Key:   aws.String("launched-by-autospotting"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("launched-for-asg"),
							Value: aws.String("mygroup"),
						},
					},
				},
				},
			},
		},
		{
			name: "create run instances input with simple LC",
			inst: instance{
				asg: &autoScalingGroup{
					name: "mygroup",
					Group: &autoscaling.Group{
						LaunchConfigurationName: aws.String("myLC"),
						LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
							LaunchTemplateId:   aws.String("lt-id"),
							LaunchTemplateName: aws.String("lt-name"),
						},
					},
					launchConfiguration: &launchConfiguration{
						LaunchConfiguration: &autoscaling.LaunchConfiguration{
							BlockDeviceMappings:      nil,
							InstanceMonitoring:       nil,
							UserData:                 aws.String("userdata"),
							AssociatePublicIpAddress: nil,
						},
					},
				},
				Instance: &ec2.Instance{
					EbsOptimized: aws.Bool(true),

					IamInstanceProfile: &ec2.IamInstanceProfile{
						Arn: aws.String("profile-arn"),
					},

					ImageId:      aws.String("ami-123"),
					InstanceType: aws.String("t2.medium"),
					KeyName:      aws.String("mykey"),

					Placement: &ec2.Placement{
						Affinity: aws.String("foo"),
					},

					SecurityGroups: []*ec2.GroupIdentifier{
						{
							GroupName: aws.String("foo"),
							GroupId:   aws.String("sg-123"),
						},
						{
							GroupName: aws.String("bar"),
							GroupId:   aws.String("sg-456"),
						},
					},

					SubnetId: nil,
				},
			}, args: args{
				instanceType: "t2.small",
				price:        1.5,
			},
			want: &ec2.RunInstancesInput{

				EbsOptimized: aws.Bool(true),

				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Arn: aws.String("profile-arn"),
				},

				ImageId: aws.String("ami-123"),

				InstanceMarketOptions: &ec2.InstanceMarketOptionsRequest{
					MarketType: aws.String("spot"),
					SpotOptions: &ec2.SpotMarketOptions{
						MaxPrice: aws.String("1.5"),
					},
				},

				InstanceType: aws.String("t2.small"),
				KeyName:      aws.String("mykey"),

				LaunchTemplate: &ec2.LaunchTemplateSpecification{
					LaunchTemplateId:   aws.String("lt-id"),
					LaunchTemplateName: aws.String("lt-name"),
				},

				MaxCount: aws.Int64(1),
				MinCount: aws.Int64(1),

				Placement: &ec2.Placement{
					Affinity: aws.String("foo"),
				},

				SecurityGroupIds: []*string{
					aws.String("sg-123"),
					aws.String("sg-456"),
				},

				SubnetId: nil,

				TagSpecifications: []*ec2.TagSpecification{{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("LaunchConfigurationName"),
							Value: aws.String("myLC"),
						},
						{
							Key:   aws.String("launched-by-autospotting"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("launched-for-asg"),
							Value: aws.String("mygroup"),
						},
					},
				},
				},
				UserData: aws.String("userdata"),
			},
		},

		{
			name: "create run instances input with full launch configuration",
			inst: instance{
				asg: &autoScalingGroup{
					name: "mygroup",
					Group: &autoscaling.Group{
						LaunchConfigurationName: aws.String("myLC"),
						LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
							LaunchTemplateId:   aws.String("lt-id"),
							LaunchTemplateName: aws.String("lt-name"),
						},
					},
					launchConfiguration: &launchConfiguration{
						LaunchConfiguration: &autoscaling.LaunchConfiguration{
							InstanceMonitoring: &autoscaling.InstanceMonitoring{
								Enabled: aws.Bool(true),
							},

							BlockDeviceMappings: []*autoscaling.BlockDeviceMapping{
								{
									DeviceName: aws.String("foo"),
								},
							},
							AssociatePublicIpAddress: aws.Bool(true),
							UserData:                 aws.String("userdata"),
						},
					},
				},
				Instance: &ec2.Instance{
					EbsOptimized: aws.Bool(true),

					IamInstanceProfile: &ec2.IamInstanceProfile{
						Arn: aws.String("profile-arn"),
					},

					ImageId:      aws.String("ami-123"),
					InstanceType: aws.String("t2.medium"),
					KeyName:      aws.String("mykey"),

					Placement: &ec2.Placement{
						Affinity: aws.String("foo"),
					},

					SecurityGroups: []*ec2.GroupIdentifier{
						{
							GroupName: aws.String("foo"),
							GroupId:   aws.String("sg-123"),
						},
						{
							GroupName: aws.String("bar"),
							GroupId:   aws.String("sg-456"),
						},
					},

					SubnetId: aws.String("subnet-123"),
				},
			}, args: args{
				instanceType: "t2.small",
				price:        1.5,
			},
			want: &ec2.RunInstancesInput{
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					{
						DeviceName: aws.String("foo"),
					},
				},

				EbsOptimized: aws.Bool(true),

				IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
					Arn: aws.String("profile-arn"),
				},

				ImageId: aws.String("ami-123"),

				InstanceMarketOptions: &ec2.InstanceMarketOptionsRequest{
					MarketType: aws.String("spot"),
					SpotOptions: &ec2.SpotMarketOptions{
						MaxPrice: aws.String("1.5"),
					},
				},

				InstanceType: aws.String("t2.small"),
				KeyName:      aws.String("mykey"),

				LaunchTemplate: &ec2.LaunchTemplateSpecification{
					LaunchTemplateId:   aws.String("lt-id"),
					LaunchTemplateName: aws.String("lt-name"),
				},

				MaxCount: aws.Int64(1),
				MinCount: aws.Int64(1),

				Monitoring: &ec2.RunInstancesMonitoringEnabled{
					Enabled: aws.Bool(true),
				},

				Placement: &ec2.Placement{
					Affinity: aws.String("foo"),
				},

				NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
					{
						AssociatePublicIpAddress: aws.Bool(true),
						DeviceIndex:              aws.Int64(0),
						SubnetId:                 aws.String("subnet-123"),
						Groups: []*string{
							aws.String("sg-123"),
							aws.String("sg-456"),
						},
					},
				},

				TagSpecifications: []*ec2.TagSpecification{{
					ResourceType: aws.String("instance"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("LaunchConfigurationName"),
							Value: aws.String("myLC"),
						},
						{
							Key:   aws.String("launched-by-autospotting"),
							Value: aws.String("true"),
						},
						{
							Key:   aws.String("launched-for-asg"),
							Value: aws.String("mygroup"),
						},
					},
				},
				},
				UserData: aws.String("userdata"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.inst.createRunInstancesInput(tt.args.instanceType, tt.args.price); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("instance.createRunInstancesInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_instance_isReadyToAttach(t *testing.T) {
	//now := time.Now()
	tenMinutesAgo := time.Now().Add(-10 * time.Minute)

	tests := []struct {
		name     string
		instance instance
		asg      *autoScalingGroup
		want     bool
	}{

		{
			name: "pending instance",
			instance: instance{
				Instance: &ec2.Instance{
					InstanceId: aws.String("i-123"),
					LaunchTime: &tenMinutesAgo,
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNamePending),
					},
				},
			},
			asg: &autoScalingGroup{
				name: "my-asg",
				Group: &autoscaling.Group{
					HealthCheckGracePeriod: aws.Int64(3600),
				},
			},
			want: false,
		},
		{
			name: "not-ready running instance",
			instance: instance{
				Instance: &ec2.Instance{
					InstanceId: aws.String("i-123"),
					LaunchTime: &tenMinutesAgo,
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNameRunning),
					},
				},
			},
			asg: &autoScalingGroup{
				name: "my-asg",
				Group: &autoscaling.Group{
					HealthCheckGracePeriod: aws.Int64(3600),
				},
			},
			want: false,
		},
		{
			name: "ready running instance",
			instance: instance{
				Instance: &ec2.Instance{
					InstanceId: aws.String("i-123"),
					LaunchTime: &tenMinutesAgo,
					State: &ec2.InstanceState{
						Name: aws.String(ec2.InstanceStateNameRunning),
					},
				},
			},
			asg: &autoScalingGroup{
				name: "my-asg",
				Group: &autoscaling.Group{
					HealthCheckGracePeriod: aws.Int64(300),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.instance.isReadyToAttach(tt.asg); got != tt.want {
				t.Errorf("instance.isReadyToAttach() = %v, want %v", got, tt.want)
			}
		})
	}
}
