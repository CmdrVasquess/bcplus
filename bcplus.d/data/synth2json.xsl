<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
				xmlns:s="https://fractalqb.de/ed/synthesis/2.4">
  <xsl:output method="text" encoding="utf8"/>
  <xsl:strip-space elements="*"/>

  <xsl:template match="/">
	<xsl:text>[</xsl:text>
	<xsl:apply-templates select="//s:synth"/>
	<xsl:text>]</xsl:text>
  </xsl:template>

  <xsl:template match="s:synth">
	<xsl:if test="position()&gt;1">
	  <xsl:text>,</xsl:text>
	</xsl:if>
	<xsl:text>{"Name":"</xsl:text>
	<xsl:value-of select="@id"/>
	<xsl:text>","Improves":"</xsl:text>
	<xsl:value-of select="@improves"/>
	<xsl:text>","Levels":[</xsl:text>
	<xsl:for-each select="s:quality">
	  <xsl:sort select="@level" data-type="number"/>
	  <xsl:if test="position()&gt;1">
		<xsl:text>,</xsl:text>
	  </xsl:if>
	  <xsl:text>{"Bonus":"</xsl:text>
	  <xsl:choose>
		<xsl:when test="@bonus">
		  <xsl:value-of select="@bonus"/>
		</xsl:when>
		<xsl:otherwise>
		  <xsl:text>–</xsl:text>
		</xsl:otherwise>
	  </xsl:choose>
	  <xsl:text>","Demand":{</xsl:text>
	  <xsl:for-each select="s:use">
		<xsl:if test="position()&gt;1">
		  <xsl:text>,</xsl:text>
		</xsl:if>
		<xsl:text>"</xsl:text>
		<xsl:value-of select="@material"/>
		<xsl:text>":</xsl:text>
		<xsl:choose>
		  <xsl:when test="@count">
			<xsl:value-of select="@count"/>
		  </xsl:when>
		  <xsl:otherwise>
			<xsl:text>1</xsl:text>
		  </xsl:otherwise>
		</xsl:choose>
	  </xsl:for-each>
	  <xsl:text>}}</xsl:text>
	</xsl:for-each>
	<xsl:text>]}
</xsl:text>
  </xsl:template>
</xsl:stylesheet>
