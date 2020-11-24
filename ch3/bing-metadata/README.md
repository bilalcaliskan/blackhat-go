## Parsing Document Metadata With Bing Scraping

As we stressed in the Shodan section, relatively benign information—when viewed in the correct context—can prove to be 
critical, increasing the likelihood that your attack against an organization succeeds. Information such as employee 
names, phone numbers, email addresses, and client software versions are often the most highly regarded because they 
provide concrete or actionable information that attackers can directly exploit or use to craft attacks that are more 
effective and highly targeted. One such source of information, popularized by a tool named [FOCA, is document metadata.](https://github.com/ElevenPaths/FOCA)

Applications store arbitrary information within the structure of a file saved to disk. In some cases, this can include 
geographical coordinates, application versions, operating system information, and usernames. Better yet, search 
engines contain advanced query filters that allow you to retrieve specific files for an organization. The remainder 
of this chapter focuses on building a tool that scrapes—or as my lawyer calls it, indexes—Bing search results to 
retrieve a target organization’s Microsoft Office documents, subsequently extracting relevant metadata.

### _Setting Up Your Environment And Planning_
Before diving into the specifics, we’ll start by stating the objectives. First, you’ll focus solely on `Office Open XML 
documents—those ending in xlsx, docx, pptx, and so on`. Although you could certainly include legacy Office data types, 
the binary formats make them exponentially more complicated, increasing code complexity and reducing readability. 
The same can be said for working with PDF files. Also, the code you develop won’t handle Bing pagination, instead 
only parsing initial page search results. We encourage you to build this into your working example and explore file 
types beyond Open XML.

`Why not just use the Bing Search APIs for building this, rather than doing HTML scraping?` Because you already know 
how to build clients that interact with structured APIs. There are practical use cases for scraping HTML pages, 
particularly when no API exists. Rather than rehashing what you already know, we’ll take this as an opportunity to 
introduce a new method of extracting data. You’ll use an excellent package, `goquery`, which mimics the functionality 
of jQuery, a JavaScript library that includes an intuitive syntax to traverse HTML documents and select data within. Start by installing `goquery`:
```shell script
$ go get github.com/PuerkitoBio/goquery
```
We have used `mod` as dependency manager, so check the [go.mod file](go.mod) for installed dependencies. 
Fortunately, that’s the only prerequisite software needed to complete the development. You’ll use standard Go 
packages to interact with Open XML files. These files, despite their file type suffix, are ZIP archives that, when 
extracted, contain XML files. The metadata is stored in two files within the docProps directory of the archive:
```shell script
$ unzip test.xlsx
$ tree
--snip--
|---docProps
|   |---app.xml
|   |---core.xml
--snip—
```

The `core.xml` file contains the author information as well as modification details. It’s structured as follows:
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata
/core-properties"
                   xmlns:dc="http://purl.org/dc/elements/1.1/"
                   xmlns:dcterms="http://purl.org/dc/terms/"
                   xmlns:dcmitype="http://purl.org/dc/dcmitype/"
                   xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <dc:creator>Dan Kottmann</dc:creator>❶
    <cp:lastModifiedBy>Dan Kottmann</cp:lastModifiedBy>❷
    <dcterms:created xsi:type="dcterms:W3CDTF">2016-12-06T18:24:42Z</dcterms:created>
    <dcterms:modified xsi:type="dcterms:W3CDTF">2016-12-06T18:25:32Z</dcterms:modified>
</cp:coreProperties>
```
The creator ❶ and lastModifiedBy ❷ elements are of primary interest. These fields contain employee or usernames that you 
can use in a social-engineering or password-guessing campaign. 

The `app.xml` file contains details about the application type and version used to create the Open XML document. 
Here’s its structure:
```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"
            xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
    <Application>Microsoft Excel</Application>❶
    <DocSecurity>0</DocSecurity>
    <ScaleCrop>false</ScaleCrop>
    <HeadingPairs>
        <vt:vector size="2" baseType="variant">
            <vt:variant>
                <vt:lpstr>Worksheets</vt:lpstr>
            </vt:variant>
            <vt:variant>
                <vt:i4>1</vt:i4>
            </vt:variant>
        </vt:vector>
    </HeadingPairs>
    <TitlesOfParts>
        <vt:vector size="1" baseType="lpstr">
            <vt:lpstr>Sheet1</vt:lpstr>
        </vt:vector>
    </TitlesOfParts>
    <Company>ACME</Company>❷
    <LinksUpToDate>false</LinksUpToDate>
    <SharedDoc>false</SharedDoc>
    <HyperlinksChanged>false</HyperlinksChanged>
    <AppVersion>15.0300</AppVersion>❸
</Properties>
```
You’re primarily interested in just a few of those elements: Application ❶, Company ❷, and AppVersion ❸. The version 
itself doesn’t obviously correlate to the Office version name, such as Office 2013, Office 2016, and so on, but a 
logical mapping does exist between that field and the more readable, commonly known alternative. The code you develop 
will maintain this mapping.

### _Defining the metadata Package_
