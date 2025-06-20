//  This file is part of the Eliona project.
//  Copyright © 2025 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eliona

import (
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
)

var CHECK_TYPE_EXTERNAL = "external" // managed by app (limits are not detected by Eliona)
var CHECK_TYPE_LIMITS = "limits"     // managed by Eliona

func CreateStatusAlarm(assetID int32) (int32, error) {
	alarmRule, _, err := client.NewClient().AlarmRulesAPI.
		PutAlarmRule(client.AuthenticationContext()).
		AlarmRule(api.AlarmRule{
			AssetId:   assetID,
			Subtype:   api.SUBTYPE_STATUS,
			Attribute: "status",
			Priority:  api.AlarmPriority(2),
			Subject:   *api.NewNullableString(api.PtrString("App Name error")),
			Message: map[string]any{
				"come": map[string]string{
					"de": "App Name meldet einen Fehlerzustand, es können Inkonsistenzen auftreten. Kontaktieren Sie den App-Entwickler, wenn das Problem weiterhin besteht.",
					"en": "App Name reports an error state, inconsistencies may occur. Contact app developer if the issue persists.",
					"fr": "App Name signale un état d'erreur, des incohérences peuvent survenir. Contactez le développeur de l'application si le problème persiste.",
					"it": "App Name segnala uno stato di errore, potrebbero verificarsi incoerenze. Contattare lo sviluppatore dell'app se il problema persiste.",
				},
				"gone": nil,
			},
			Tags:                []string{},
			Enable:              api.PtrBool(true),
			CheckType:           *api.NewNullableString(&CHECK_TYPE_LIMITS),
			RequiresAcknowledge: api.PtrBool(false),
			Urldoc:              *api.NewNullableString(api.PtrString("https://doc.eliona.io/collection/eliona-english/eliona-apps/apps/app-name")),
			High:                *api.NewNullableFloat64(api.PtrFloat64(1)),
		}).
		Execute()
	if err != nil {
		return 0, err
	}
	return alarmRule.GetId(), nil
}
