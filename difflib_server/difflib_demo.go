// Note: Vuln to SSRF, so internal use only

package main

import (
	"bufio"
	"fmt"
	"github.com/003random/difflib"
	"github.com/gorilla/mux"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var templateString = `
<!doctype html>
<html>

<head>
  <meta charset="utf-8" />
  <title>File diff</title>
  <link rel="stylesheet" href="https://use.fontawesome.com/releases/v5.6.3/css/all.css"
    integrity="sha384-UHRtZLI+pbxtHCWp1t77Bi1L4ZtiqrqD80Kn4Z8NTSRyMA2Fd33n5dQ8lWUE00s/" crossorigin="anonymous">
  <link rel="stylesheet" href="http://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.8/styles/default.min.css">
  <style type="text/css">
  .hljs {
    display: inline;
    overflow-x: none;
    padding: 0;
    background: white;
  }

  .diff-table {
    display: block;
    overflow-x: auto;
    margin: 5vw;
    font-family: Console, Liberation Mono, DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New;
    background-color: white;
    border-collapse: collapse;
    font-size: 0.9em;
    border: 1px solid #ddd;
    border-radius: 3px;
    margin-bottom: 16px;
    margin-top: 16px;
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

  .collapse-icon {
    cursor: pointer;
    padding: 10px;
    margin-left: 15px;
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
</style>
</head>
<form action="#" method="POST">
    <input type="text" name="first" id="first" value="https://varanid.io/static/test.txt">
    <input type="text" name="second" id="second" value="https://varanid.io/static/test1.txt">
    trim:<input type="hidden" name="trim" value="0"><input type="checkbox" onclick="this.previousSibling.value=1-this.previousSibling.value">
    <button type="submit">diff</button>
</form>

{{.Diff}}

<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.0/jquery.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.15.8/highlight.min.js"></script>
<script>
  hljs.initHighlightingOnLoad();
  $(document).ready(function () {

    $(".diff-table > tbody > tr:not(.table-header):not(.new-part)").mouseenter(function () {
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

    $(".diff-table > tbody > tr:not(.table-header):not(.new-part)").mouseleave(function () {
      $(this).css({
        "background": "white",
        "background-color": "white"
      });
      $(this).find(".hljs").css({
        "background": "white",
        "background-color": "white"
      });

      $(this).find(".code").children(".fa-copy").remove();
    });

    $(".collapse-icon").on("click", function () {
      $(this).toggleClass("fa-chevron-down").toggleClass("fa-chevron-right");
      $(".diff-table > tbody > tr:not(.table-header)").toggle();
      $(".table-header").children().eq(1).css("width", "100%");
    });

    var switched = false;
    $(".line-num").on("click", function () {
      if (!switched) {
        $(".code").each(function () {
          var html = "<td class='" + $(this).attr("class") + "'>" + $(this).html(); + "</td>"
          console.log(html);
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

  function copyStringToClipboard(str) {
    // Create new element
    var el = document.createElement('textarea');
    // Set value (string to be copied)
    el.value = str;
    // Set non-editable to avoid focus and move outside of view
    el.setAttribute('readonly', '');
    el.style = {
      position: 'absolute',
      left: '-9999px'
    };
    document.body.appendChild(el);
    // Select text inside element
    el.select();
    // Copy text to clipboard
    document.execCommand('copy');
    // Remove temporary element
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
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

    trim, _ := strconv.ParseBool(r.PostForm["trim"][0])
	  diff := difflib.Diff(getResponseAsSlice(r.PostForm["first"][0]), getResponseAsSlice(r.PostForm["second"][0]), trim)
		htmlDiff = difflib.HTMLDiff(diff, "Difference")
	}

	tmpl, _ := template.New("diffTemplate").Parse(templateString)
	err := tmpl.Execute(w, map[string]interface{}{
		"Diff": template.HTML(htmlDiff),
	})
	if err != nil {
		log.Print(err)
	}
}

func getResponseAsSlice(url string) []string {
	fmt.Println(url)
	var client http.Client
	resp, err := client.Get(url)
	if err != nil {
		log.Print(err)
		return []string{}
	}
	defer resp.Body.Close()

	s := bufio.NewScanner(resp.Body)
	body := []string{}
	for s.Scan() {
		body = append(body, html.EscapeString(s.Text()))
	}
	if err := s.Err(); err != nil {
		log.Print(err)
		return []string{}
	}
	return body
}
