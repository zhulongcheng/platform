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

const VisualizationIconBar: SFC<Props> = ({
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
      id="Bar"
      x="0px"
      y="0px"
      viewBox="0 0 150 150"
      preserveAspectRatio="none meet"
      shapeRendering="geometricPrecision"
      className={containerClass}
    >
      <rect
        x="2"
        y="108.4"
        style={{stroke: colorA, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        width="26.8"
        height="16.6"
      />
      <rect
        x="31.8"
        y="82.4"
        style={{stroke: colorB, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        width="26.8"
        height="42.6"
      />
      <rect
        x="61.6"
        y="28.8"
        style={{stroke: colorC, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        width="26.8"
        height="96.2"
      />
      <rect
        x="91.4"
        y="47.9"
        style={{stroke: colorA, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        width="26.8"
        height="77.1"
      />
      <rect
        x="121.2"
        y="25"
        style={{stroke: colorB, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        width="26.8"
        height="100"
      />
      <rect
        x="2"
        y="108.4"
        style={{fill: colorA}}
        className="visualization-icon--fill"
        width="26.8"
        height="16.6"
      />
      <rect
        x="31.8"
        y="82.4"
        style={{fill: colorB}}
        className="visualization-icon--fill"
        width="26.8"
        height="42.6"
      />
      <rect
        x="61.6"
        y="28.8"
        style={{fill: colorC}}
        className="visualization-icon--fill"
        width="26.8"
        height="96.2"
      />
      <rect
        x="91.4"
        y="47.9"
        style={{fill: colorA}}
        className="visualization-icon--fill"
        width="26.8"
        height="77.1"
      />
      <rect
        x="121.2"
        y="25"
        style={{fill: colorB}}
        className="visualization-icon--fill"
        width="26.8"
        height="100"
      />
    </svg>
  )
}

export default VisualizationIconBar
