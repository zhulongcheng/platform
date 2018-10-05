// Libraries
import React, {SFC} from 'react'
import classnames from 'classnames'

// Types
import {ThemeColor} from 'src/clockface'

interface Props {
  colorA: ThemeColor
  colorB: ThemeColor
  colorC: ThemeColor
  colorGrey: ThemeColor
  highlight?: boolean
  className?: string
  stroke: number
}

const VisualizationIconGauge: SFC<Props> = ({
  colorA,
  colorB,
  colorC,
  colorGrey,
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
      id="VisualizationGauge"
      x="0px"
      y="0px"
      viewBox="0 0 150 150"
      preserveAspectRatio="none meet"
      shapeRendering="geometricPrecision"
      className={containerClass}
    >
      <g>
        <path
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          d="M110.9,110.9c19.9-19.9,19.9-52,0-71.9s-52-19.9-71.9,0s-19.9,52,0,71.9"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="39.1"
          y1="110.9"
          x2="35"
          y2="115"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="110.9"
          y1="110.9"
          x2="115"
          y2="115"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="122"
          y1="94.5"
          x2="127.2"
          y2="96.6"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="125.8"
          y1="75"
          x2="131.5"
          y2="75"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="122"
          y1="55.5"
          x2="127.2"
          y2="53.4"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="110.9"
          y1="39.1"
          x2="115"
          y2="35"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="94.5"
          y1="28"
          x2="96.6"
          y2="22.8"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="75"
          y1="24.2"
          x2="75"
          y2="18.5"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="55.5"
          y1="28"
          x2="53.4"
          y2="22.8"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="39.1"
          y1="39.1"
          x2="35"
          y2="35"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="28"
          y1="55.5"
          x2="22.8"
          y2="53.4"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="24.2"
          y1="75"
          x2="18.5"
          y2="75"
        />
        <line
          style={{stroke: colorGrey, strokeWidth: `${stroke}px`}}
          className="visualization-icon--line nuetral"
          x1="28"
          y1="94.5"
          x2="22.8"
          y2="96.6"
        />
      </g>
      <path
        style={{fill: colorGrey}}
        className="visualization-icon--fill nuetral"
        d="M78.6,73.4L75,56.3l-3.6,17.1c-0.2,0.5-0.3,1-0.3,1.6c0,2.2,1.8,3.9,3.9,3.9s3.9-1.8,3.9-3.9C78.9,74.4,78.8,73.9,78.6,73.4z"
      />
      <path
        style={{fill: colorA}}
        className="visualization-icon--fill"
        d="M58.9,58.9c8.9-8.9,23.4-8.9,32.3,0l17.1-17.1c-18.4-18.4-48.2-18.4-66.5,0C32.5,50.9,27.9,63,27.9,75h24.2C52.2,69.2,54.4,63.3,58.9,58.9z"
      />
      <path
        style={{stroke: colorA, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        d="M58.9,58.9c8.9-8.9,23.4-8.9,32.3,0l17.1-17.1c-18.4-18.4-48.2-18.4-66.5,0C32.5,50.9,27.9,63,27.9,75h24.2C52.2,69.2,54.4,63.3,58.9,58.9z"
      />
      <path
        style={{fill: colorB}}
        className="visualization-icon--fill"
        d="M58.9,91.1c-4.5-4.5-6.7-10.3-6.7-16.1H27.9c0,12,4.6,24.1,13.8,33.3L58.9,91.1z"
      />
      <path
        style={{stroke: colorB, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        d="M58.9,91.1c-4.5-4.5-6.7-10.3-6.7-16.1H27.9c0,12,4.6,24.1,13.8,33.3L58.9,91.1z"
      />
      <path
        style={{fill: colorC}}
        className="visualization-icon--fill"
        d="M91.1,91.1l17.1,17.1c18.4-18.4,18.4-48.2,0-66.6L91.1,58.9C100.1,67.8,100.1,82.2,91.1,91.1z"
      />
      <path
        style={{stroke: colorC, strokeWidth: `${stroke}px`}}
        className="visualization-icon--line"
        d="M91.1,91.1l17.1,17.1c18.4-18.4,18.4-48.2,0-66.6L91.1,58.9C100.1,67.8,100.1,82.2,91.1,91.1z"
      />
    </svg>
  )
}

export default VisualizationIconGauge
