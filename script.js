/*global d3:false, EventSource:false */
(function () {
  'use strict';

  // TODO: it might be better to rotate the whole thing, and then individually
  // rotate each piece of text back

  // Returns a value for the SVG "transform" property for translating by x, y.
  var _translate = function (x, y) {
    return ['translate(', x, ',', y, ')'].join('');
  };

  var setter = function (params, func) {
    if (func === undefined) {
      func = 'attr';
    }
    return function (node) {
      d3.entries(params).forEach(function (item) {
        node[func](item.key, item.value);
      });
      return node;
    };
  };

  var Orientation = {
    wide: function (data, axisLength, barLength) {
      return {
        axis: {orient: 'left', transform: ''},
        range: {
          x: [0, axisLength - axisLength / data.length],
          y: [0, barLength]
        },
        svg: {width: barLength + 105, height: axisLength + 65},
        label: {
          x: -35,
          y: (axisLength - axisLength / data.length) * 0.5,
          rotate: -90
        },
        translate: function (x) {
          return function (d) {
            return _translate(0, x(d.x));
          };
        },
        bar: function (x, y, dx) {
          return setter({
            x: 0,
            y: 1,
            height: Math.max(1, dx - 1),
            width: function (d) { return y(d.y); }
          });
        },
        text: function (x, y, dx) {
          return setter({
            x: function (d) {
              var len = this.getComputedTextLength();
              return y(d.y) > len + 30 ? y(d.y) - 6 - len : y(d.y) + 6;
            },
            y: dx * 0.5,
            'class': function (d) {
              var len = this.getComputedTextLength();
              return y(d.y) > len + 30 ? 'inside' : 'outside';
            },
            'dominant-baseline': 'middle'
          });
        }
      };
    },

    tall: function (data, axisLength, barLength) {
      return {
        axis: {orient: 'bottom', transform: _translate(0, barLength)},
        range: {
          x: [0, axisLength - axisLength / data.length],
          y: [barLength, 0]
        },
        svg: {width: axisLength + 65, height: barLength + 105},
        label: {
          x: (axisLength - axisLength / data.length) * 0.5,
          y: barLength + 50,
          rotate: 0
        },
        translate: function (x, y) {
          return function (d) {
            return _translate(x(d.x), y(d.y));
          };
        },
        bar: function (x, y, dx) {
          return setter({
            x: 1,
            y: 0,
            height: function (d) { return barLength - y(d.y); },
            width: Math.max(1, dx - 1)
          });
        },
        text: function (x, y, dx) {
          return setter({
            x: dx * 0.5,
            y: function (d) {
              return y(d.y) > barLength - 30 ? -6 : 18;
            },
            'class': function (d) {
              return y(d.y) > barLength - 30 ? 'outside' : 'inside';
            },
            'text-anchor': 'middle'
          });
        }
      };
    }
  };

  var histogram = function (data, opts) {

    if (data.length <= 1) {
      // TODO: don't return, show something
      return;
    }

    if (opts.Label) {
      document.title = opts.Label;
    }

    // TODO: dynamic with screen resize, underscore debounce?
    var orient = Orientation[opts.Wide ? 'wide' : 'tall'](data, 500, 500);

    var x = d3.scale.linear()
      .domain([d3.min(data, function (d) { return d.x; }),
               d3.max(data, function (d) { return d.x; }) + opts.Bucket])
      .range(orient.range.x);

    var y = d3.scale.linear()
      .domain([0, d3.max(data, function (d) { return d.y; })])
      .range(orient.range.y);

    var dx = (x(1) - x(0)) * opts.Bucket;

    var axis = d3.svg.axis().scale(x).orient(orient.axis.orient);

    var svg = d3.select('body').append('svg')
      .attr('width', orient.svg.width)
      .attr('height', orient.svg.height)
      .append('g')
      .attr('transform', _translate(50, 50));
      // TODO translate amount based on axis width & svg width

    svg.append('g')
      .attr('transform', _translate(orient.label.x, orient.label.y))
      .append('text')
      .text(opts.Label)
      .attr('class', 'label')
      .attr('text-anchor', 'middle')
      .attr('font-size', '1.1em')
      .attr('font-weight', 'bold')
      .attr('transform', 'rotate(' + orient.label.rotate + ')');

    var bar = svg.selectAll('.bar').data(data)
      .enter()
      .append('g')
      .attr('class', 'bar')
      .attr('transform', orient.translate(x, y, dx));

    orient.bar(x, y, dx)(
      bar.append('rect'));
    orient.text(x, y, dx)(
      bar.append('text').text(function (d) { return d.y; }));

    svg.append('g')
      .attr('class', 'axis')
      .attr('transform', orient.axis.transform)
      .call(axis);
  };

  var pushHistogram = window.pushHistogram = function (data) {
    var hist = d3.map(data.Values).entries().map(function (i) {
      return {x: parseFloat(i.key), y: i.value};
    });
    hist.sort(d3.ascending);
    d3.select('svg').remove();
    histogram(hist, data);
  };

  var events = new EventSource('/data');
  events.onmessage = function (e) {
    var data = JSON.parse(e.data);
    if (data.type && data.type === 'error') {
      // TODO show the error
      return;
    }
    // TODO show EOF, others
    // TODO switch on "histogram" type
    pushHistogram(data);
  };
})();
