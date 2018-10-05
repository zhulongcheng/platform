// Libraries
import React, {SFC} from 'react'
import classnames from 'classnames'

// Types
import {ThemeColor} from 'src/clockface'

interface Props {
  colorA: ThemeColor
  colorB: ThemeColor
  colorC: ThemeColor
  highlight?: boolean
  className?: string
  stroke: number
}

const VisualizationIconStacked: SFC<Props> = ({
  colorA,
  colorB,
  colorC,
  highlight,
  className,
  stroke,
}) => {
  const containerClass = classnames('visualization-icon', {
    highlight,
    [`${className}`]: className,
  })

  return (
    <svg
      width="100%"
      height="100%"
      version="1.1"
      id="LineStacked"
      x="0px"
      y="0px"
      viewBox="0 0 150 150"
      preserveAspectRatio="none meet"
      shapeRendering="geometricPrecision"
      className={containerClass}
    >
      <polygon
        style={{fill: colorA}}
        className="visualization-icon--fill"
        points="148,25 111.5,25 75,46 38.5,39.1 2,85.5 2,125 148,125 	"
      />
      <polyline
        style={{stroke: colorA, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        points="2,85.5 38.5,39.1 75,46 111.5,25 148,25 	"
      />
      <polygon
        style={{fill: colorB}}
        className="visualization-icon--fill"
        points="148,53 111.5,49.9 75,88.5 38.5,71 2,116 2,125 148,125 	"
      />
      <polyline
        style={{stroke: colorB, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        points="2,116 38.5,71 75,88.5 111.5,49.9 148,53 	"
      />
      <polygon
        style={{fill: colorC}}
        className="visualization-icon--fill"
        points="148,86.2 111.5,88.6 75,108.6 38.5,98 2,121.1 2,125 148,125 	"
      />
      <polyline
        style={{stroke: colorC, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        points="2,121.1 38.5,98 75,108.6 111.5,88.6 148,86.2 	"
      />
    </svg>
  )
}

export default VisualizationIconStacked
