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

const VisualizationIconStepPlot: SFC<Props> = ({
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
      id="StepPlot"
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
        points="148,61.9 129.8,61.9 129.8,25 93.2,25 93.2,40.6 56.8,40.6 56.8,25 20.2,25 20.2,67.8 2,67.8 2,125 148,125 	"
      />
      <polyline
        style={{stroke: colorA, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        points="2,67.8 20.2,67.8 20.2,25 56.8,25 56.8,40.6 93.2,40.6 93.2,25 129.8,25 129.8,61.9 148,61.9 	"
      />
      <polygon
        style={{fill: colorB}}
        className="visualization-icon--fill"
        points="148,91.9 129.8,91.9 129.8,70.2 93.2,70.2 93.2,67 56.8,67 56.8,50.1 20.2,50.1 20.2,87 2,87 2,125 148,125 	"
      />
      <polyline
        style={{stroke: colorB, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        points="2,87 20.2,87 20.2,50.1 56.8,50.1 56.8,67 93.2,67 93.2,70.2 129.8,70.2 129.8,91.9 148,91.9 	"
      />
      <polygon
        style={{fill: colorC}}
        className="visualization-icon--fill"
        points="148,103.5 129.8,103.5 129.8,118.2 93.2,118.2 93.2,84.5 56.8,84.5 56.8,75 20.2,75 20.2,100.2 2,100.2 2,125 148,125 	"
      />
      <polyline
        style={{stroke: colorC, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        points="2,100.2 20.2,100.2 20.2,75 56.8,75 56.8,84.5 93.2,84.5 93.2,118.2 129.8,118.2 129.8,103.5 148,103.5 	"
      />
    </svg>
  )
}

export default VisualizationIconStepPlot
