<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform" version="1.0"
				xmlns:st="https://fractalqb.de/ed/ship-types/3.0">

  <xsl:output method="html" encoding="utf-8"/>
  <xsl:strip-space elements="*"/>

  <xsl:template match="st:ship-types">
	<html lang="de">
	  <head>
		<meta charset="utf-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
		<link rel="stylesheet" href="s/css/bcp.css"/>
		<!-- >>> local-style >>> -->
  		<style>
main table {
  border-collapse: collapse;
}
main table th {
  border: 2px solid var(--colBkg);
}
main table th.shdr { font-size: 60%; }
main table th.sortable-asc {
	background: linear-gradient(to bottom,var(--mnuBkg),var(--colFrg));
}
main table th.sortable-dsc {
	background: linear-gradient(to bottom,var(--colFrg),var(--mnuBkg));
}
main table td {
  border: 2px solid var(--colBkg);
  text-align: center;
}
main td.rstr { color: var(--colGood); }
main td:nth-child(7), td:nth-child(12), td:nth-child(19), td:nth-child(24) {
  border-right: 2px solid var(--colSec);
}
main td:first-child, td:nth-child(2) {
  text-align: left;
}
main td:nth-child(3), td:nth-child(6), td:nth-child(7) {
  text-align: right;
}
main tbody tr:hover {
  background-color: #993000;
}
  		</style>
		<script src="s/jqui/external/jquery/jquery.js"></script>
  		<script src="s/jqui/jquery-ui.min.js"></script>
  		<script src="s/js/sortable.js"></script>
  	  </head>
 	  <body>
  		<main>
  		  <table>
			<thead>
  			  <tr>
				<th rowspan="2"
					onclick="sortable(this, 0, cmpStr)">Ship</th>
				<th rowspan="2"
					onclick="sortable(this, 1, cmpStr)">Manufacturer</th>
				<th rowspan="2"
					onclick="sortable(this, 2, cmpNum)">Cost [Cr]</th>
				<th rowspan="2"
					onclick="sortable(this, 3, cmpStr)"
					title="Large / Medium / Small Ship">Sz</th>
				<th rowspan="2"
					onclick="sortable(this, 4, cmpNum)"
					title="Number Crew Members">C</th>
				<th rowspan="2"
					onclick="sortable(this, 5, cmpNum)">AG</th>
				<th rowspan="2"
					onclick="sortable(this, 6, cmpNum)">HR</th>
  				<th colspan="5">Maximum</th>
  				<th colspan="7">Core Modules</th>
  				<th colspan="5">Hardpoints/Utils</th>
  				<th colspan="9">Optional Modules</th>
  			  </tr>
  			  <tr>
				<th class="shdr"
					onclick="sortable(this, 7, cmpNum)"
					title="Speed">Spd</th>
				<th class="shdr"
					onclick="sortable(this, 8, cmpNum)" title="Boost">Bst</th>
				<th class="shdr"
					onclick="sortable(this, 9, cmpNum)"
					title="Jump Range">Jmp</th>
				<th class="shdr"
					onclick="sortable(this, 10, cmpNum)" title="Cargo">Crg</th>
				<th class="shdr"
					onclick="sortable(this, 11, cmpNum)"
					title="Passengers">Pgr</th>
				<th class="shdr"
					onclick="sortable(this, 12, cmpNum)"
					title="Power Plant">PP</th>
				<th class="shdr"
					onclick="sortable(this, 13, cmpNum)"
					title="Thrusters">TH</th>
				<th class="shdr"
					onclick="sortable(this, 14, cmpNum)"
					title="Frame Shift Drive">FS</th>
				<th class="shdr"
					onclick="sortable(this, 15, cmpNum)"
					title="Life Support">LS</th>
				<th class="shdr"
					onclick="sortable(this, 16, cmpNum)"
					title="Power Distributor">PD</th>
				<th class="shdr"
					onclick="sortable(this, 17, cmpNum)" title="Scanner">SC</th>
				<th class="shdr"
					onclick="sortable(this, 18, cmpNum)"
					title="Fuel Tank">FT</th>
				<th class="shdr"
					onclick="sortable(this, 19, cmpNum)"
					title="Small Hardpoints">S</th>
				<th class="shdr"
					onclick="sortable(this, 20, cmpNum)"
					title="Medium Hardpoints">M</th>
				<th class="shdr"
					onclick="sortable(this, 21, cmpNum)"
					title="Large Hardpoints">L</th>
				<th class="shdr"
					onclick="sortable(this, 22, cmpNum)"
					title="Huge Hardpoints">H</th>
				<th class="shdr"
					onclick="sortable(this, 23, cmpNum)"
					title="Utility Mounts">U</th>
				<th class="shdr"
					onclick="sortable(this, 24, cmpStr)"
					title="Supports Fighter Hangar">F</th>
				<th class="shdr" onclick="sortable(this, 25, cmpStr)">1</th>
				<th class="shdr" onclick="sortable(this, 26, cmpStr)">2</th>
				<th class="shdr" onclick="sortable(this, 27, cmpStr)">3</th>
				<th class="shdr" onclick="sortable(this, 28, cmpStr)">4</th>
				<th class="shdr" onclick="sortable(this, 29, cmpStr)">5</th>
				<th class="shdr" onclick="sortable(this, 30, cmpStr)">6</th>
				<th class="shdr" onclick="sortable(this, 31, cmpStr)">7</th>
				<th class="shdr" onclick="sortable(this, 32, cmpStr)">8</th>
  			  </tr>
			</thead>
			<tbody>
			  <xsl:apply-templates select="st:ship"/>
			</tbody>
		  </table>
  		</main>
		<script>
$("main table tbody").sortable();
		</script>
	  </body>
	</html>
  </xsl:template>
		  
  <xsl:template match="st:ship">
	<tr id="{@jname}">
	  <td><xsl:value-of select="@name"/></td>
	  <td><xsl:value-of select="@manf"/></td>
	  <td><xsl:value-of select="format-number(@price,'#,###.#')"/></td>
	  <td><xsl:choose>
		<xsl:when test="@size='small'">S</xsl:when>
		<xsl:when test="@size='medium'">M</xsl:when>
		<xsl:when test="@size='large'">L</xsl:when>
		<xsl:otherwise>???</xsl:otherwise>
	  </xsl:choose></td>
	  <td><xsl:value-of select="@crew"/></td>
	  <td><xsl:value-of select="st:feature[@type='agility']"/></td>
	  <td><xsl:value-of select="st:feature[@type='hardness']"/></td>
	  <td><xsl:value-of select="st:feature[@type='max-speed']"/></td>
	  <td><xsl:value-of select="st:feature[@type='max-boost']"/></td>
	  <td><xsl:value-of select="st:feature[@type='max-jump']"/></td>
	  <td><xsl:value-of select="st:feature[@type='max-cargo']"/></td>
	  <td><xsl:value-of select="st:feature[@type='max-passengers']"/></td>
	  <xsl:apply-templates select="st:core[@module='powerplant']"/>
	  <xsl:apply-templates select="st:core[@module='thrusters']"/>
	  <xsl:apply-templates select="st:core[@module='fsd']"/>
	  <xsl:apply-templates select="st:core[@module='life-support']"/>
	  <xsl:apply-templates select="st:core[@module='power-distributor']"/>
	  <xsl:apply-templates select="st:core[@module='scanner']"/>
	  <xsl:apply-templates select="st:core[@module='fuel-tank']"/>
	  <td><xsl:value-of select="st:hardpoints/@small"/></td>
	  <td><xsl:value-of select="st:hardpoints/@medium"/></td>
	  <td><xsl:value-of select="st:hardpoints/@large"/></td>
	  <td><xsl:value-of select="st:hardpoints/@huge"/></td>
	  <td><xsl:value-of select="st:utilitiy/@mounts"/></td>
	  <td><xsl:if test="@fighter='true'">✔</xsl:if></td>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=1]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=2]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=3]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=4]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=5]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=6]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=7]"/>
	  </xsl:call-template>
	  <xsl:call-template name="internal">
		<xsl:with-param name="mods" select="st:internal[@size=8]"/>
	  </xsl:call-template>
	</tr>
  </xsl:template>

  <xsl:template match="st:core">
	<td><xsl:value-of select="@class"/></td>
  </xsl:template>

  <xsl:template name="internal">
	<xsl:param name="mods"/>
	<xsl:variable name="modno" select="count($mods)"/>
	<xsl:choose>
	  <xsl:when test="$modno=0">
		<td></td>
	  </xsl:when>
	  <xsl:when test="$modno=1">
		<xsl:variable name="mod" select="$mods[1]"/>
		<xsl:choose>
		  <xsl:when test="$mod/@restriction">
			<xsl:variable name="rlb">
			  <xsl:call-template name="rlabel">
				<xsl:with-param name="mods" select="$mods"/>
			  </xsl:call-template>
			</xsl:variable>
			<td class="rstr"
				title="{$rlb}"><xsl:value-of select="$mod/@count"/></td>
		  </xsl:when>
		  <xsl:otherwise>
			<td><xsl:value-of select="$mod/@count"/></td>
		  </xsl:otherwise>
		</xsl:choose>
	  </xsl:when>
	  <xsl:otherwise>
		<xsl:variable name="sum">
		  <xsl:call-template name="rsum">
			<xsl:with-param name="sum" select="0"/>
			<xsl:with-param name="mods" select="$mods"/>
		  </xsl:call-template>
		</xsl:variable>
		<xsl:variable name="rlb">
		  <xsl:call-template name="rlabel">
			<xsl:with-param name="mods" select="$mods"/>
		  </xsl:call-template>
		</xsl:variable>
		<td class="rstr"
			title="{$rlb}"><xsl:value-of select="$sum"/></td>
	  </xsl:otherwise>
	</xsl:choose>
  </xsl:template>

  <xsl:template name="rsum">
	<xsl:param name="sum"/>
	<xsl:param name="mods"/>
	<xsl:choose>
	  <xsl:when test="count($mods)=0">
		<xsl:value-of select="$sum"/>
	  </xsl:when>
	  <xsl:otherwise>
		<xsl:variable name="s">
		  <xsl:call-template name="rsum">
			<xsl:with-param name="sum" select="$sum"/>
			<xsl:with-param name="mods" select="$mods[position()&gt;1]"/>
		  </xsl:call-template>
		</xsl:variable>
		<xsl:value-of select="$mods[1]/@count + $s"/>		
	  </xsl:otherwise>
	</xsl:choose>
  </xsl:template>
  
  <xsl:template name="rlabel">
	<xsl:param name="mods"/>
	<xsl:variable name="lbstd">
	  <xsl:text>S:</xsl:text>
	  <xsl:choose>
		<xsl:when test="$mods[not(@restriction)]">
		  <xsl:value-of select="$mods[not(@restriction)]/@count"/>
		</xsl:when>
		<xsl:otherwise>-</xsl:otherwise>
	  </xsl:choose>
	</xsl:variable>
	<xsl:variable name="lbmil">
	  <xsl:text>M:</xsl:text>
	  <xsl:choose>
		<xsl:when test="$mods[@restriction='military']">
		  <xsl:value-of select="$mods[@restriction='military']/@count"/>
		</xsl:when>
		<xsl:otherwise>-</xsl:otherwise>
	  </xsl:choose>
	</xsl:variable>
	<xsl:variable name="lbtrs">
	  <xsl:text>T:</xsl:text>
	  <xsl:choose>
		<xsl:when test="$mods[@restriction='tourist']">
		  <xsl:value-of select="$mods[@restriction='tourist']/@count"/>
		</xsl:when>
		<xsl:otherwise>-</xsl:otherwise>
	  </xsl:choose>
	</xsl:variable>
	<xsl:value-of select="$lbstd"/>
	<xsl:text> </xsl:text>
	<xsl:value-of select="$lbmil"/>
	<xsl:text> </xsl:text>
	<xsl:value-of select="$lbtrs"/>
  </xsl:template>
  
</xsl:stylesheet>
