package handler

import "testing"

var d1empty = `
	var bgGreen="#78ff47";
	var bgYellow="#fbff79";
	var bgOrange="#ffb23b";
	var bgRed="#ff5047";

			document.getElementById("D1-p-001").innerHTML="Nema upozorenja.";	
			
			document.getElementById("D1-p-002").innerHTML="Nema upozorenja.";	
			
			document.getElementById("D1-p-003").innerHTML="Nema upozorenja.";
`

var d1emptyResult = `var bgGreen="#78ff47";
var bgYellow="#fbff79";
var bgOrange="#ffb23b";
var bgRed="#ff5047";
document.getElementById("D1-p-001").innerHTML="Nema upozorenja.";
document.getElementById("D1-p-002").innerHTML="Nema upozorenja.";
document.getElementById("D1-p-003").innerHTML="Nema upozorenja.";
`

var d2empty = `
var bgGreen="#78ff47";
var bgYellow="#fbff79";
var bgOrange="#ffb23b";
var bgRed="#ff5047";

document.getElementById("p-001").innerHTML="Nema upozorenja.";
document.getElementById("p-002").innerHTML="Nema upozorenja.";
document.getElementById("p-003").innerHTML="Nema upozorenja.";
`

func TestHandlerMeteoalarmNormScriptContent_RemovesWhitespacesAndNewlines(t *testing.T) {
	input := `var bgGreen="#78ff47"; 
var bgYellow="#fbff79"; 
	var bgOrange="#ffb23b"; var bgRed="#ff5047";`

	expected := `var bgGreen="#78ff47";
var bgYellow="#fbff79";
var bgOrange="#ffb23b";
var bgRed="#ff5047";
`

	result := normScriptContent(input)
	if result != expected {
		t.Errorf("expected [%s], got [%s]", expected, result)
	}
}

func TestHandlerMeteoalarmNormScriptContent_EmptyAlert(t *testing.T) {
	result := normScriptContent(d1empty)
	if result != d1emptyResult {
		t.Errorf("expected [%s], got [%s]", d1empty, result)
	}
}

func TestHandlerMeteoalarmNormScriptContent_RealAlert_One(t *testing.T) {
	var input = `
var bgGreen="#78ff47";
        var bgYellow="#fbff79";
        var bgOrange="#ffb23b";
        var bgRed="#ff5047";
document.getElementById("D1-001-wind").setAttribute("class","active");  document.getElementsByClassName("D1-poly_ME001")[0].style.fill=bgOrange;document.getElementsByClassName("D1-alarm-text-ME001")[0].children[0].style.borderColor=bgOrange;document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora) </li>";    document.getElementById("D1-001-rain").setAttribute("class","active");document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Kiša-najmanje 20mm za 12h  </li>";    document.getElementById("D1-002-wind").setAttribute("class","active");  document.getElementsByClassName("D1-poly_ME002")[0].style.fill=bgOrange;document.getElementsByClassName("D1-alarm-text-ME002")[0].children[0].style.borderColor=bgOrange;document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";     document.getElementById("D1-002-rain").setAttribute("class","active");  document.getElementsByClassName("D1-poly_ME002")[0].style.fill=bgOrange;document.getElementsByClassName("D1-alarm-text-ME002")[0].children[0].style.borderColor=bgOrange;document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";      document.getElementById("D1-002-thunderstorm").setAttribute("class","active");document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>"; document.getElementById("D1-003-wind").setAttribute("class","active");  document.getElementsByClassName("D1-poly_ME003")[0].style.fill=bgOrange;document.getElementsByClassName("D1-alarm-text-ME003")[0].children[0].style.borderColor=bgOrange;document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";     document.getElementById("D1-003-rain").setAttribute("class","active");  document.getElementsByClassName("D1-poly_ME003")[0].style.fill=bgOrange;document.getElementsByClassName("D1-alarm-text-ME003")[0].children[0].style.borderColor=bgOrange;document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";      document.getElementById("D1-003-thunderstorm").setAttribute("class","active");document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>"; document.getElementById("D1-003-coastalevent").setAttribute("class","active");document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Talasasto more. Daglasova skala  4 do 5</li>";
`

	var expected = `var bgGreen="#78ff47";
var bgYellow="#fbff79";
var bgOrange="#ffb23b";
var bgRed="#ff5047";
document.getElementById("D1-001-wind").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME001")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME001")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora) </li>";
document.getElementById("D1-001-rain").setAttribute("class","active");
document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Kiša-najmanje 20mm za 12h  </li>";
document.getElementById("D1-002-wind").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME002")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME002")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";
document.getElementById("D1-002-rain").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME002")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME002")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";
document.getElementById("D1-002-thunderstorm").setAttribute("class","active");
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("D1-003-wind").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME003")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME003")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";
document.getElementById("D1-003-rain").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME003")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME003")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";
document.getElementById("D1-003-thunderstorm").setAttribute("class","active");
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("D1-003-coastalevent").setAttribute("class","active");
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Talasasto more. Daglasova skala  4 do 5</li>";
`

	result := normScriptContent(input)
	if result != expected {
		t.Errorf("expected [%s], got [%s]", expected, result)
	}
}

func TestHandlerMeteoalarmNormScriptContent_RealAlert_Two(t *testing.T) {
	var input = `
var bgGreen="#78ff47";
        var bgYellow="#fbff79";
        var bgOrange="#ffb23b";
        var bgRed="#ff5047";


                                document.getElementById("001-wind").setAttribute("class","active");
                                document.getElementsByClassName("poly_ME001")[0].style.fill=bgRed;
                                document.getElementsByClassName("alarm-text-ME001")[0].children[0].style.borderColor=bgRed;
                                document.getElementById("p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Udari vjetra >25 m/s (>90km/h ili oko 10 Bofora) </li>";

                                document.getElementById("001-rain").setAttribute("class","active");document.getElementById("p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h  </li>";
                                document.getElementById("001-thunderstorm").setAttribute("class","active");document.getElementById("p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000. </li>";
                                document.getElementById("002-wind").setAttribute("class","active");
                                document.getElementsByClassName("poly_ME002")[0].style.fill=bgRed;
                                document.getElementsByClassName("alarm-text-ME002")[0].children[0].style.borderColor=bgRed;
                                document.getElementById("p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Udari vjetra >25 m/s (>90km/h ili oko 10 Bofora)</li>";

                                document.getElementById("002-rain").setAttribute("class","active");
                                document.getElementsByClassName("poly_ME002")[0].style.fill=bgRed;
                                document.getElementsByClassName("alarm-text-ME002")[0].children[0].style.borderColor=bgRed;
                                document.getElementById("p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Jaka kiša-najmanje 20mm za 3h </li>";

                                document.getElementById("002-thunderstorm").setAttribute("class","active");document.getElementById("p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
                                document.getElementById("003-wind").setAttribute("class","active");
                                document.getElementsByClassName("poly_ME003")[0].style.fill=bgRed;
                                document.getElementsByClassName("alarm-text-ME003")[0].children[0].style.borderColor=bgRed;
                                document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Udari vjetra >25 m/s (>90km/h ili oko 10 Bofora)</li>";

                                document.getElementById("003-rain").setAttribute("class","active");
                                document.getElementsByClassName("poly_ME003")[0].style.fill=bgRed;
                                document.getElementsByClassName("alarm-text-ME003")[0].children[0].style.borderColor=bgRed;
                                document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Jaka kiša-najmanje 20mm za 3h </li>";

                                document.getElementById("003-thunderstorm").setAttribute("class","active");document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
                                document.getElementById("003-coastalevent").setAttribute("class","active");document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Talasasto more. Daglasova skala  4 do 5</li>";
`

	var expected = `var bgGreen="#78ff47";
var bgYellow="#fbff79";
var bgOrange="#ffb23b";
var bgRed="#ff5047";
document.getElementById("001-wind").setAttribute("class","active");
document.getElementsByClassName("poly_ME001")[0].style.fill=bgRed;
document.getElementsByClassName("alarm-text-ME001")[0].children[0].style.borderColor=bgRed;
document.getElementById("p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Udari vjetra >25 m/s (>90km/h ili oko 10 Bofora) </li>";
document.getElementById("001-rain").setAttribute("class","active");
document.getElementById("p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h  </li>";
document.getElementById("001-thunderstorm").setAttribute("class","active");
document.getElementById("p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000. </li>";
document.getElementById("002-wind").setAttribute("class","active");
document.getElementsByClassName("poly_ME002")[0].style.fill=bgRed;
document.getElementsByClassName("alarm-text-ME002")[0].children[0].style.borderColor=bgRed;
document.getElementById("p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Udari vjetra >25 m/s (>90km/h ili oko 10 Bofora)</li>";
document.getElementById("002-rain").setAttribute("class","active");
document.getElementsByClassName("poly_ME002")[0].style.fill=bgRed;
document.getElementsByClassName("alarm-text-ME002")[0].children[0].style.borderColor=bgRed;
document.getElementById("p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Jaka kiša-najmanje 20mm za 3h </li>";
document.getElementById("002-thunderstorm").setAttribute("class","active");
document.getElementById("p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("003-wind").setAttribute("class","active");
document.getElementsByClassName("poly_ME003")[0].style.fill=bgRed;
document.getElementsByClassName("alarm-text-ME003")[0].children[0].style.borderColor=bgRed;
document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Udari vjetra >25 m/s (>90km/h ili oko 10 Bofora)</li>";
document.getElementById("003-rain").setAttribute("class","active");
document.getElementsByClassName("poly_ME003")[0].style.fill=bgRed;
document.getElementsByClassName("alarm-text-ME003")[0].children[0].style.borderColor=bgRed;
document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: red;\"></i>Jaka kiša-najmanje 20mm za 3h </li>";
document.getElementById("003-thunderstorm").setAttribute("class","active");
document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("003-coastalevent").setAttribute("class","active");
document.getElementById("p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Talasasto more. Daglasova skala  4 do 5</li>";
`
	result := normScriptContent(input)
	if result != expected {
		t.Errorf("expected [%s], got [%s]", expected, result)
	}
}

func TestHandlerMeteoalarmRemoveWhitespacesAndNewlines(t *testing.T) {
	var input = `var bgGreen="#78ff47";
var bgYellow="#fbff79";
var bgOrange="#ffb23b";
var bgRed="#ff5047";
document.getElementById("D1-001-wind").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME001")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME001")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora) </li>";
document.getElementById("D1-001-rain").setAttribute("class","active");
document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Kiša-najmanje 20mm za 12h  </li>";
document.getElementById("D1-002-wind").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME002")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME002")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";
document.getElementById("D1-002-rain").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME002")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME002")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";
document.getElementById("D1-002-thunderstorm").setAttribute("class","active");
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("D1-003-wind").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME003")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME003")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";
document.getElementById("D1-003-rain").setAttribute("class","active");
document.getElementsByClassName("D1-poly_ME003")[0].style.fill=bgOrange;
document.getElementsByClassName("D1-alarm-text-ME003")[0].children[0].style.borderColor=bgOrange;
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";
document.getElementById("D1-003-thunderstorm").setAttribute("class","active");
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("D1-003-coastalevent").setAttribute("class","active");
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Talasasto more. Daglasova skala  4 do 5</li>";
`

	var expected = `document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora) </li>";
document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Kiša-najmanje 20mm za 12h  </li>";
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";
document.getElementById("D1-p-002").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)</li>";
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Jaka kiša-najmanje 10mm za 3h </li>";
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Moguće vremenske nepogode i grad. CAPE index oko 1000.</li>";
document.getElementById("D1-p-003").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: yellow;\"></i>Talasasto more. Daglasova skala  4 do 5</li>";
`

	result := removeUselessLines(input)
	if result != expected {
		t.Errorf("expected [%s], got [%s]", expected, result)
	}
}

func TestHandlerMeteoalarmExtractAlerts_DayOne(t *testing.T) {
	input := `document.getElementById("D1-p-001").innerHTML+="<li><i class=\"fa fa-warning\" style=\"color: orange;\"></i>Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora) </li>";`

	expected := Alert{
		Region: "001",
		Level:  "orange",
		Text:   "Udari vjetra >17 m/s (>50km/h ili oko 8 Bofora)",
	}

	result := extractAlert(input)
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

}

//func TestNormScriptContent_EmptyString(t *testing.T) {
//	input := ""
//	expected := ""
//	result := normScriptContent(input)
//	if result != expected {
//		t.Errorf("expected %s, got %s", expected, result)
//	}
//}
//
//func TestNormScriptContent_OnlyWhitespacesAndNewlines(t *testing.T) {
//	input := " \n \n "
//	expected := ""
//	result := normScriptContent(input)
//	if result != expected {
//		t.Errorf("expected %s, got %s", expected, result)
//	}
//}
//
//func TestNormScriptContent_NoWhitespacesOrNewlines(t *testing.T) {
//	input := "varbgGreen=\"#78ff47\";varbgYellow=\"#fbff79\";"
//	expected := "varbgGreen=\"#78ff47\";varbgYellow=\"#fbff79\";"
//	result := normScriptContent(input)
//	if result != expected {
//		t.Errorf("expected %s, got %s", expected, result)
//	}
//}
