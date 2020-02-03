xquery version "1.0-ml";
declare namespace x="http://www.w3.org/1999/xhtml";
declare default function namespace "local";

(:
to run:
curl -H'Content-Type:text' 'http://ml-utils-qa.springernature.app/ml-utils/v1/eval.xqy' -d "@scripts/gocd-pipelines.xqy"
:)

declare function server($name as xs:string) {
  let $host := $name || ".gocd.tools.engineering"
  let $html := xdmp:tidy(xdmp:http-get("http://" || $host || "/go/pipelines")[2])[2]
  for $g in $html//x:div[@class = "content_wrapper_inner"]
  return group($g, $host)
};

declare function group($html, $host) {
  let $name := $html/x:h2/fn:string()
  where $name != "Release.Engineering"
  return
    for $p in ($html/x:div[@class = "pipeline"])
    return pipeline($p, $host, $name)
};

declare function pipeline($html, $host, $group) {
   let $name := fn:normalize-space($html//x:h3/x:a/fn:string())
   let $paused := fn:exists($html//x:span[@class = "paused_by"])
   let $time := fn:substring(($html//x:div[@class = "schedule_time"]/@title)[1], 14, 12)
   let $last-run := try { xs:date(fn:substring($time, 9, 4) || "-" || month(fn:substring($time, 4, 3)) || "-" || fn:substring($time, 1, 2)) } catch * { xs:date("1900-01-01")}
   where fn:not($paused) and $last-run > fn:current-date() - xs:yearMonthDuration("P3Y")
   return <pipeline name="{$name}" group="{$group}" host="{$host}" last-run="{$last-run}"/>
};

declare function month($s) {
  let $i := fn:index-of(("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"), $s)
  return fn:string(if ($i < 10) then "0" || $i else $i)
};

declare function csv($p) {
  fn:string-join(($p/@name, $p/@group, $p/@host, $p/@last-run), ",")
};

"Pipeline Name,Pipeline Group,GoCD Server,Last Run",
server(("05")) ! csv(.)
