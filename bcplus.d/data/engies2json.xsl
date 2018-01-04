<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform" version="1.0"
				xmlns:qe="https://fractalqb.de/ed/engineers/2.4">

  <xsl:output method="text" encoding="utf-8"/>
  <xsl:strip-space elements="*"/>

  <xsl:template match="/qe:engineers">
	<xsl:text>{"groups": [</xsl:text>
	<xsl:apply-templates select="qe:group"/>
	<xsl:text>], "modules": [</xsl:text>
    <xsl:apply-templates select="qe:module"/>
	<xsl:text>], "engineers": [</xsl:text>
    <xsl:apply-templates select="qe:engineer"/>
    <xsl:text>]}</xsl:text>
  </xsl:template>

  <xsl:template match="qe:group">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>{"@": "</xsl:text>
	<xsl:value-of select="@key"/>
	<xsl:text>", </xsl:text>
	<xsl:apply-templates select="qe:name"/>
	<xsl:text>}</xsl:text>
  </xsl:template>
  
  <xsl:template match="qe:module">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>{"@": "</xsl:text>
	<xsl:value-of select="@key"/>
	<xsl:text>", "group": "</xsl:text>
	<xsl:value-of select="@group"/>
	<xsl:text>", "blueprints": [</xsl:text>
	<xsl:apply-templates select="qe:blueprint"/>
	<xsl:text>]</xsl:text>
	<xsl:if test="qe:name">
	  <xsl:text>, "names": {</xsl:text>
	  <xsl:apply-templates select="qe:name"/>
	  <xsl:text>}</xsl:text>
	</xsl:if>
	<xsl:text>}</xsl:text>
  </xsl:template>

  <xsl:template match="qe:blueprint">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>{</xsl:text>
	<xsl:text>"@": "</xsl:text>
	<xsl:value-of select="@key"/>
	<xsl:text>", "grade": </xsl:text>
	<xsl:value-of select="@grade"/>
	<xsl:text>, "effects": {</xsl:text>
	<xsl:apply-templates select="qe:effect"/>
	<xsl:text>}, "materials": [</xsl:text>
	<xsl:apply-templates select="qe:material"/>
	<xsl:text>]</xsl:text>
	<xsl:if test="qe:name">
	  <xsl:text>, "names": {</xsl:text>
	  <xsl:apply-templates select="qe:name"/>
	  <xsl:text>}</xsl:text>
	</xsl:if>
	<xsl:text>}</xsl:text>
  </xsl:template>

  <xsl:template match="qe:effect">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>"</xsl:text>
	<xsl:value-of select="@property"/>
	<xsl:text>": {"min": </xsl:text>
	<xsl:value-of select="@min"/>
	<xsl:text>, "max": </xsl:text>
	<xsl:value-of select="@max"/>
	<xsl:text>}</xsl:text>
  </xsl:template>

  <xsl:template match="qe:material">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>{"@": "</xsl:text>
	<xsl:value-of select="@key"/>
	<xsl:text>", "count": </xsl:text>
	<xsl:choose>
	  <xsl:when test="@count">
		<xsl:value-of select="@count"/>
	  </xsl:when>
	  <xsl:otherwise>
		<xsl:text>1</xsl:text>
	  </xsl:otherwise>
	</xsl:choose>
	<xsl:text>}</xsl:text>
  </xsl:template>
  
  <xsl:template match="qe:engineer">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>{"name": "</xsl:text>
	<xsl:value-of select="@name"/>
	<xsl:text>"</xsl:text>
	<xsl:if test="qe:discover">
	  <xsl:text>, "discover": [</xsl:text>
	  <xsl:for-each select="qe:discover">
		<xsl:if test="position() &gt; 1">
		  <xsl:text>, </xsl:text>
		</xsl:if>
		<xsl:text>"</xsl:text>
		<xsl:value-of select="."/>
		<xsl:text>"</xsl:text>
	  </xsl:for-each>
	  <xsl:text>]</xsl:text>
	</xsl:if>
	<xsl:if test="qe:requirement">
	  <xsl:text>, "reqirement": "</xsl:text>
	  <xsl:value-of select="normalize-space(qe:requirement)"/>
	  <xsl:text>"</xsl:text>
	</xsl:if>
	<xsl:if test="qe:unlock">
	  <xsl:text>, "unlock": "</xsl:text>
	  <xsl:value-of select="normalize-space(qe:unlock)"/>
	  <xsl:text>"</xsl:text>
	</xsl:if>
	<xsl:if test="qe:rep">
	  <xsl:text>, "rep": [</xsl:text>
	  <xsl:for-each select="qe:rep">
		<xsl:if test="position() &gt; 1">
		  <xsl:text>, </xsl:text>
		</xsl:if>
		<xsl:text>"</xsl:text>
		<xsl:value-of select="normalize-space(.)"/>
		<xsl:text>"</xsl:text>
	  </xsl:for-each>
	  <xsl:text>]</xsl:text>
	</xsl:if>
	<xsl:text>, "mods": {</xsl:text>
	<xsl:apply-templates select="qe:modification"/>
	<xsl:text>}</xsl:text>
	<xsl:text>}</xsl:text>
  </xsl:template>

  <xsl:template match="qe:modification">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>"</xsl:text>
	<xsl:value-of select="@module"/>
	<xsl:text>": </xsl:text>
	<xsl:value-of select="@grade"/>
  </xsl:template>
  
  <xsl:template match="qe:name">
	<xsl:if test="position() &gt; 1">
	  <xsl:text>, </xsl:text>
	</xsl:if>
	<xsl:text>"</xsl:text>
	<xsl:choose>
	  <xsl:when test="@lang">
		<xsl:value-of select="@lang"/>
	  </xsl:when>
	  <xsl:otherwise>
		<xsl:text>default</xsl:text>
	  </xsl:otherwise>
	</xsl:choose>
	<xsl:text>": "</xsl:text>
	<xsl:value-of select="."/>
	<xsl:text>"</xsl:text>
  </xsl:template>
  
</xsl:stylesheet>
