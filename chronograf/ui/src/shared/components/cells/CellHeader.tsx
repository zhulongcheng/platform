import React, {SFC} from 'react'
import classnames from 'classnames'
import {isCellUntitled} from 'src/dashboards/utils/cellGetters'

interface Props {
  isEditable: boolean
  cellName: string
}

const LayoutCellHeader: SFC<Props> = ({isEditable, cellName}) => {
  const headerClass = classnames('cell--header', {
    'cell--draggable cell--header-draggable': isEditable,
  })

  const nameClass = classnames('cell--name', {
    'cell--name__default': isCellUntitled(cellName),
  })

  return (
    <div className={headerClass}>
      <span className={nameClass}>{cellName}</span>
    </div>
  )
}

export default LayoutCellHeader
