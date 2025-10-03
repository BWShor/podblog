-- inline_html.lua
-- Handle @@html:start ... @@html:end blocks as inline or block raw HTML

local collecting = false
local buffer = {}

-- Collect inline tokens
function Str(el)
  if el.text == "@@html:start" then
    collecting = true
    buffer = {}
    return {}
  elseif el.text == "@@html:end" then
    collecting = false
    local raw_html = table.concat(buffer, " ")
    buffer = {}
    return pandoc.RawInline("html", raw_html)
  elseif collecting then
    table.insert(buffer, el.text)
    return {}
  end
end

function Space(el)
  if collecting then
    table.insert(buffer, " ")
    return {}
  end
end

function SoftBreak(el)
  if collecting then
    table.insert(buffer, "\n")
    return {}
  end
end

function LineBreak(el)
  if collecting then
    table.insert(buffer, "\n")
    return {}
  end
end

-- Post-process: if a paragraph contains only one RawInline("html"),
-- turn it into RawBlock("html") so no <p> wrapper is added.
function Para(el)
  if #el.content == 1 and el.content[1].t == "RawInline"
     and el.content[1].format == "html" then
    return pandoc.RawBlock("html", el.content[1].text)
  end
  return el
end
