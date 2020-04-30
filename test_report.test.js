const expect = require('@jest/globals').expect
require('./test_report.js')

const mockData = [{
  "TestResults": [{
    TestName: "my_sample_test",
    Package: "test/package",
    Output: [
      "test output 1\n",
      "test output 2\n",
      "test output 3\n",
    ]
  }]
}]


test('test testGroupListHandler', () => {
  const testResultsElem = document.createElement('div')
  testResultsElem.id =  'testResults'
  const testGroupListElem = document.createElement('div')
  testGroupListElem.id =  'testGroupList'
  document.body.insertAdjacentElement('beforeend', testResultsElem)
  document.body.insertAdjacentElement('beforeend', testGroupListElem)
  const goTestReport = window.GoTestReport.init(document)
  const divElem = document.createElement('div')
  const testIdAttr = document.createAttribute('data-testid')
  testIdAttr.value = "0"
  const indexAttr = document.createAttribute('data-index')
  indexAttr.value = "0"
  divElem.attributes.setNamedItem(testIdAttr)
  divElem.attributes.setNamedItem(indexAttr)
  goTestReport.testGroupListHandler(divElem, mockData)
  const testOutputDiv = divElem.querySelector('div.testOutput')
  expect(testOutputDiv).toBeDefined()

  const consoleElem = testOutputDiv.querySelector('.console.failed')
  expect(consoleElem).toBeDefined()
  expect(consoleElem.textContent).toBe('test output 1\ntest output 2\ntest output 3\n')

  const testDetailElem = testOutputDiv.querySelector('.testDetail')
  expect(testDetailElem).toBeDefined()

  const packageElem = testDetailElem.querySelector('.package')
  expect(packageElem).toBeDefined()
  expect(packageElem.innerHTML).toBe(`<strong>Package:</strong> test/package`)

  const filenameElem = testDetailElem.querySelector('.filename')
  expect(filenameElem).toBeDefined()
  expect(filenameElem.innerHTML).toBe(`<strong>Filename:</strong> `)
});