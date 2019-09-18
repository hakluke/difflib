// Note: Vuln to SSRF, so internal use only

package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"github.com/003random/difflib"
	"github.com/gorilla/mux"
	"github.com/yosssi/gohtml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
  "strings"
  
  jsbeautifier "github.com/003random/jsbeautifier-go/jsbeautifier"
)

var templateString = `
<!doctype html>
<html>

<head>
  <meta charset="utf-8" />
  <title>File diff</title>
  <link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.6.3/css/all.css"
    integrity="sha384-UHRtZLI+pbxtHCWp1t77Bi1L4ZtiqrqD80Kn4Z8NTSRyMA2Fd33n5dQ8lWUE00s/" crossorigin="anonymous">
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.8/styles/default.min.css">
  <style type="text/css">
  .hljs {
    display: inline;
    overflow-x: none;
    padding: 0;
    background: white;
  }

  .diff-table {
    max-height: 80vh;
    display: block;
    overflow-x: auto;
    font-family: Console, Liberation Mono, DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New;
    background-color: white;
    border-collapse: collapse;
    font-size: 0.9em;
    border: 1px solid #ddd;
    border-radius: 3px;
    margin-bottom: 16px;
    margin-top: 16px;
    -webkit-box-shadow: 0 0 3px gray;
    box-shadow: 0 0 3px gray;
  }

  .diff-table>tbody>tr {
    line-height: 0px;
    height: 26px !important;
  }

  .diff-table>tbody>tr>.code {
    font-size: 12px;
    width: 100%;
  }

  .diff-table>tbody>tr>td {
    overflow: hidden;
  }

  .line-num:hover {
    color: rgba(27, 31, 35, .6);
  }

  .line-num {
    cursor: pointer;
    font-size: 0.8em;
    padding: 5px;
    padding-left: 15px;
    padding-right: 15px;
    color: rgba(27, 31, 35, .3);
    background-color: #fcfcfc;
  }

  .line-num-added {
    background-color: #cdffd8 !important;
  }

  .line-num-deleted {
    background-color: #ffdce0 !important;
  }

  .line-num .added {
    background-color: #cdffd8;
  }

  .line-num .deleted {
    background-color: #cdffd8;
  }

  .code {
    padding-right: 20px;
    padding-left: 20px;
    border-color: #bef5cb;
  }

  .added,
  .added>pre>code {
    background-color: #e6ffed !important;
  }

  .deleted,
  .deleted>pre>code {
    background-color: #ffeef0 !important;
  }

  .table-header {
    background-color: #fafbfc;
    height: 32px;
    border-bottom: 1px solid #ddd;
  }

  .delta-type {
    position: relative;
    top: 12px;
    left: -10px;
  }

  .deleted-text {
    color: black !important;
    padding: 1px;
    background-color: #fdb8c0;
    border-bottom-right-radius: .2em;
    border-top-right-radius: .2em;
    border-bottom-left-radius: .2em;
    border-top-left-radius: .2em;
  }

  .added-text {
    color: black !important;
    padding: 1px;
    background-color: #acf2bd;
    border-bottom-right-radius: .2em;
    border-top-right-radius: .2em;
    border-bottom-left-radius: .2em;
    border-top-left-radius: .2em;
  }

  .line-num-added+.code>pre>code:empty {
    padding: 10px;
  }
  
  .line-num-deleted+.code>pre>code:empty {
    padding: 10px;
  }

  .fa-copy {
    position: relative;
    right: 0px;
    top: 1px;
    float: right;
    color: lightskyblue;
    border-radius: 60px;
    box-shadow: 0px 0px 2px #888;
    padding: 0.4em 0.5em;
    background-color: white;
    cursor: pointer;
  }

  .fa-copy:hover {
    transform: scale(1.1);
  }

  .new-part {
    height: 30px;
    background-color: #f6f6f6;
    width: 100%;
  }

  .new-part > td {
    text-align: center;
    color: gray;
    opacity: 0.6;
  }

  .excluded > .code > pre > code {
    background-color: #ffebd9 !important;
  }

  .excluded > .code, .excluded > td {
    background-color: rgba(255, 119, 0, 0.15) !important;
  }

  #diff-table-bottom {
    background-color: white;
  }

  #diff-table-bottom > td {
    box-shadow: 0 -5px 5px -5px black; 
    position: sticky; 
    bottom: 0;
  }

  #diff-table-bottom > td > div {
    max-width: 33vw; 
    text-align: center; 
    width: 33%; 
    line-height: 26px; 
    float: left;
  }

  .added-color {
    background-color: #e6ffed !important;
  }

  .deleted-color {
    background-color: #ffeef0 !important;
  }

  .excluded-color {
    background-color: #ffebd9 !important;
    width: 34% !important;
    max-width: 34vw !important;
  }
  
  #diff-search {
    margin-left: 10px;
    border: 1px solid white;
    width: 100%;
    height: 100%;
  }

  #diff-text {
    margin-top: 3px;
    float: left;
/*  font-weight: bold;  */
    font-size: 1.1em;
  }

  .diff-search-icon {
    cursor: pointer;
    padding: 10px;
    margin-left: 15px;
  }

  #table-search {
    display: none;
  }

  #table-search > td {
    border-bottom: 2px solid lightgray;
    height: 35px;
  }

  .re-match {
    border: 2px solid rgb(255, 0, 0, 0.4);
  }

  .search-highlight {
    background-color: rgb(255, 0, 0, 0.4);
  }

  #diff-search-results {
    display: none;
    background-color: #f9fbff;
  }

  .results-row {
    height: 20px;
  }

  .results-num-1, .results-num-2 {
    color: rgba(27, 31, 35, 0.3);
  }
  </style>
</head>

<form action="/" method="POST">
    <input type="text" name="first" id="first" value="https://gist.githubusercontent.com/003random/550d91fa4443ea8b3a9a6ab8cfb55128/raw/2f585a3f00d2c1cad18d31e646206f4b30a6176d/test">
    <input type="text" name="second" id="second" value="https://test.poc-server.com">
    trim:<input type="hidden" name="trim" value="1"><input type="checkbox" onclick="this.previousSibling.value=1-this.previousSibling.value" checked="checked">
    filter:<input type="hidden" name="filter" value="1"><input type="checkbox" onclick="this.previousSibling.value=1-this.previousSibling.value" checked="checked">
    beautify:<input type="hidden" name="beautify" value="1"><input type="checkbox" onclick="this.previousSibling.value=1-this.previousSibling.value" checked="checked">
    <button type="submit">diff</button>
</form>
{{if .AfterPost}}
  {{if .IsDifferent}}
    {{.Diff}}
  {{else}}
    Hashes were the same. No difference found.
  {{end}}
{{end}}

<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.0/jquery.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.8/highlight.min.js"></script>
<script>
hljs.initHighlightingOnLoad();
$(document).ready(function () {
  $("#diff-search").on("input", function() {
    $("#diff-search-results-table").html("");
    var value = $("#diff-search").val();
    if (value === "" || value.length < 3) {
      $(".re-match").removeClass("re-match");
      return
    }

    $(".diff-table > tbody > tr:not(.table-header):not(.new-part):not(#table-search):not(#diff-search-results)").each(function() {
      var text = $(this).children(".code").text().trim();
      if ($(this).children(".code").hasClass("added") || $(this).children(".code").hasClass("deleted")) {
        if (text[0] == '+' || text[0] == '-') {
          text = text.substr(1).trim();
        }
      }

      var match = false;
      var regex = false;
      var include = false;
      var re = new RegExp(value, "g")

      if (text.includes(value)) {
        match = true;
        include = true;
      } else if (re.test(text)) {
        match = true;
        regex = true;
      }

      if (match) {
        $(this).children("td").eq(2).addClass("re-match");
        var result = value;
        re.exec(text);
        if (regex) {
          while (match = re.exec(text)) {
            addSearchResult($(this).children("td").eq(0).text(), $(this).children("td").eq(1).text(), getValuePlusContext(match.index, match[0].length, text));
          }
        } else if (include) {
          var indexes = getIndicesOf(value, text);
          for (i = 0; i < indexes.length; i++) {
            addSearchResult($(this).children("td").eq(0).text(), $(this).children("td").eq(1).text(), getValuePlusContext(indexes[i], value.length, text));
          }
        }
      } else {
        $(this).children("td").eq(2).removeClass("re-match");
      }
    });
  });

  function getValuePlusContext(index, len, text) {
    var contextMargin = 20;
    var word = text.substr(index, len);
    var before = "";

    for (i = 0; i < contextMargin; i++) {
      if (index > i) {
        before = text.substr(index-(i+1), (i+1));
      }
    }
    after = "";
    for (i = 1; i < contextMargin+1; i++) {
      if (index + len < text.length) {
        after = text.substr(index + len, i)
      }
    }

    return (escapeHtml(before) + "<span class='search-highlight'>" + escapeHtml(word) + "</span>" + escapeHtml(after)).trim()
  }

  function addSearchResult(line1, line2, result) {
    $("#diff-search-results-table").append("<tr class='results-row'><td class='line-num line-num-normal'>"+line1+"</td><td class='line-num line-num-normal'>"+line2+"</td><td class='results-text'>" + result + "</td></tr>");      
  }

  function getIndicesOf(searchStr, str) {
    var searchStrLen = searchStr.length;
    if (searchStrLen == 0) {
      return [];
    }
    var startIndex = 0, index, indices = [];
    while ((index = str.indexOf(searchStr, startIndex)) > -1) {
      indices.push(index);
      startIndex = index + searchStrLen;
    }
    return indices;
  }

  $(".diff-search-icon").on("click", function() {
    $("#table-search, #diff-search-results").fadeToggle();
  });

  $(".diff-table > tbody > tr:not(.table-header):not(.new-part):not(#table-search):not(#diff-search-results)").mouseenter(function () {
    $(this).css({
      "background": "#f9fbff",
      "background-color": "#f9fbff"
    });
    $(this).find(".hljs").css({
      "background": "#f9fbff",
      "background-color": "#f9fbff"
    });

    $(this).find(".code").prepend("<i class=\"fa fa-copy\"></i>");
    $(".fa-copy").on("click", function () {
      copyStringToClipboard($(this).siblings("pre").children("code").text());
    });
  });

  $(".diff-table > tbody > tr:not(.table-header):not(.new-part):not(#table-search):not(#diff-search-results)").mouseleave(function () {
    $(this).css({
      "background": "white",
      "background-color": "white"
    });
    var hljs = $(this).find(".hljs")
    if (!hljs.parent().parent().hasClass("excluded")) {
      hljs.css({
        "background": "white",
        "background-color": "white"
      });
    }

    $(this).find(".code").children(".fa-copy").remove();
  });

  var switched = false;
  $(".line-num").on("click", function () {
    if (!switched) {
      $(".code").each(function () {
        var html = "<td class='" + $(this).attr("class") + "'>" + $(this).html(); + "</td>"
        if ($(this).siblings().first().hasClass("line-num-added")) {
          html = "<td></td>";
          $(this).siblings().first().removeClass("line-num-added")
        }

        $(html).insertAfter($(this).siblings(".line-num").first());
        if ($(this).siblings().first().hasClass("line-num-deleted")) {
          $(this).siblings(".line-num-deleted").eq(1).removeClass("line-num-deleted");
          $(this).siblings(".code").addClass("deleted");
          $(this).html("");
          $(this).removeClass("deleted");
        }
      });
      $(".code").css("width", "50%");
      $(".code").css("max-width", "50%");
      switched = true;
    }
  });
  $(".new-part").each(function() {
      if (!$(this).prev().hasClass("table-header")) {
          $(this).prev().css({
              "transform": "scale(1)",
              "box-shadow": "0 5px 5px -5px #333"
          });
      }
      $(this).next().css({
          "transform": "scale(1)",
          "box-shadow": "0 -5px 5px -5px #333"
      });
  });
});

var entityMap = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
  '/': '&#x2F;',
  '=': '&#x3D;'
};

function escapeHtml(string) {
  return String(string).replace(/[&<>"'=\/]/g, function (s) {
      return entityMap[s];
  });
}

function copyStringToClipboard(str) {
  var el = document.createElement('textarea');
  el.value = str;
  el.setAttribute('readonly', '');
  el.style = {
    position: 'absolute',
    left: '-9999px'
  };
  document.body.appendChild(el);
  el.select();
  document.execCommand('copy');
  document.body.removeChild(el);
}
</script>
</body>
</html>
`

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", Hello)
	http.Handle("/", r)
	fmt.Println("Starting up on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Hello(w http.ResponseWriter, r *http.Request) {
  var htmlDiff string
  var isDifferent bool
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		trim, _ := strconv.ParseBool(r.PostForm["trim"][0])
		filter, _ := strconv.ParseBool(r.PostForm["filter"][0])
		beautify, _ := strconv.ParseBool(r.PostForm["beautify"][0])

		old, isJS := getResponse(r.PostForm["first"][0])
		new, isJS := getResponse(r.PostForm["second"][0])

		isDifferent = fmt.Sprintf("%x", md5.Sum([]byte(old))) != fmt.Sprintf("%x", md5.Sum([]byte(new)))

		if isDifferent {
      diff := difflib.Diff(getResponseAsSlice(old, beautify, isJS), getResponseAsSlice(new, beautify, isJS), trim)
			if filter {
				again, isJS := getResponse(r.PostForm["second"][0])
				dynamicDiff := difflib.Diff(getResponseAsSlice(new, beautify, isJS), getResponseAsSlice(again, beautify, isJS), trim)
				dynamicLineNums := []int{}
				for _, line := range dynamicDiff {
					if line.Delta != difflib.Common.String() {
						dynamicLineNums = append(dynamicLineNums, line.Number[0])
					}
        }
				sort.Ints(dynamicLineNums)

				if len(dynamicLineNums) > 0 {
					for i, line := range diff {
						for _, num := range dynamicLineNums {
							if num > line.Number[0] {
								break
							} else if num == line.Number[0] {
								diff[i].Exclude = true
								break
							}
						}
					}
				}
			}
			htmlDiff = difflib.HTMLDiff(diff, "Difference")
    }
	}

	tmpl, _ := template.New("diffTemplate").Parse(templateString)
	err := tmpl.Execute(w, map[string]interface{}{
    "Diff": template.HTML(htmlDiff),
    "IsDifferent": isDifferent,
    "AfterPost": r.Method == http.MethodPost,
	})
	if err != nil {
		log.Print(err)
	}
}

func getResponse(url string) (string, bool) {
  var isJS bool = false
	fmt.Println(url)
	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		log.Print(err)
		return "", false
	}
  defer resp.Body.Close()
  if _, ok := resp.Header["Content-Type"]; ok && strings.Contains(resp.Header["Content-Type"][0], "javascript") {
    isJS = true
  }
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return "", isJS
	}
	return string(d), isJS
}

func getResponseAsSlice(resp string, beautify, isJS bool) []string {
	body := []string{}
  if isJS && beautify {
    fmt.Println("isjs")
    options := jsbeautifier.DefaultOptions()
    resp = jsbeautifier.Beautify(&resp, options)
  }	else if beautify {
		resp = gohtml.Format(resp)
	}

	s := bufio.NewScanner(strings.NewReader(resp))
	for s.Scan() {
		body = append(body, s.Text())
	}
	if err := s.Err(); err != nil {
		log.Print(err)
		return []string{}
	}

	return body
}
