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

const VisualizationIconLine: SFC<Props> = ({
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
      id="VisualizationLine"
      x="0px"
      y="0px"
      viewBox="0 0 150 150"
      preserveAspectRatio="none meet"
      shapeRendering="geometricPrecision"
      className={containerClass}
    >
      <g>
        <polyline
          style={{stroke: colorA, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line"
          points="2,111.8 38.5,90.8 75,25 111.5,47.2 148,40 	"
        />
        <polygon
          style={{fill: colorA}}
          className="visualization-icon--fill"
          points="148,40 111.5,47.2 75,25 38.5,90.8 2,111.8 2,125 148,125 	"
        />
      </g>
      <g>
        <polyline
          style={{stroke: colorB, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line"
          points="2,90.8 38.5,49.3 75,61.7 111.5,95.5 148,88.2 	"
        />
        <polygon
          style={{fill: colorB}}
          className="visualization-icon--fill"
          points="148,88.2 111.5,95.5 75,61.7 38.5,49.3 2,90.8 2,125 148,125 	"
        />
      </g>
      <g>
        <polyline
          style={{stroke: colorC, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line"
          points="2,115 38.5,116.5 75,85.7 111.5,106.3 148,96 	"
        />
        <polygon
          style={{fill: colorC}}
          className="visualization-icon--fill"
          points="148,96 111.5,106.3 75,85.7 38.5,116.5 2,115 2,125 148,125 	"
        />
      </g>
    </svg>
  )
}

export default VisualizationIconLine
