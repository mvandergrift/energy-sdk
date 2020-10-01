// Copyright (C) 2020 Ramon Quitales
//
// This file is part of go-health-mate-sdk.
//
// go-health-mate-sdk is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-health-mate-sdk is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-health-mate-sdk.  If not, see <http://www.gnu.org/licenses/>.

package healthmate

import "golang.org/x/oauth2"

// HealthMateEndpoint is the endpoints for Withings Health Mate
var HealthMateEndpoint = oauth2.Endpoint{
	AuthURL:  "https://account.withings.com/oauth2_user/authorize2",
	TokenURL: "https://account.withings.com/oauth2/token",
}
