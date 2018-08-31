import _ from 'lodash'
import {FluxTable, DygraphValue} from 'src/types'

export const fluxTablesToDygraph = (data: FluxTable[]): DygraphValue[][] => {
  interface V {
    [time: string]: number[]
  }

  const valuesForTime: V = {}

  data.forEach(table => {
    const header = table.data[0]
    const timeColIndex = header.findIndex(col => col === '_time')

    table.data.slice(1).forEach(row => {
      valuesForTime[row[timeColIndex]] = Array(data.length).fill(null)
    })
  })

  data.forEach((table, i) => {
    const header = table.data[0]
    const timeColIndex = header.findIndex(col => col === '_time')
    const valueColIndexMap = {}
    header
      .filter(
        el =>
          el !== '_time' &&
          el !== '_start' &&
          el !== '_stop' &&
          el !== '_field' &&
          el !== 'table' &&
          el !== 'result' &&
          el !== '' &&
          !(el in table.groupKey)
      )
      .forEach(h => {
        valueColIndexMap[h] = header.findIndex(col => col === h)
      })

    table.data.slice(1).forEach(row => {
      const time = row[timeColIndex]
      const values = []
      Object.keys(valueColIndexMap).forEach(val => {
        const value = row[valueColIndexMap[val]]
        values.push(+value)
      })
      valuesForTime[time] = values
    })
  })

  return _.sortBy(Object.entries(valuesForTime), ([time]) => time).map(
    ([time, values]) => [new Date(time), ...values]
  )
}
