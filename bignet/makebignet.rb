
def makeLinkFile(src, neighbors, link_counts)
	filename = "#{src}.lnx"
    port = 6000 + src
	service = "localhost:#{port}"

	failed = false

	File.open(filename, "w") do |f|
		begin
			f.puts(service)

			neighbors.each do |n|
                neighbor_port = 6000 + n
				neighbor_service = "localhost:#{neighbor_port}"

				(1..link_counts).each do |i|
					src_link = "#{src}.#{n}.0.#{i}"
					dst_link = "#{n}.#{src}.0.#{i}"
					f.puts("#{neighbor_service} #{src_link} #{dst_link}")
				end
			end
		rescue
			failed = true
		end
	end

	File.delete(filename) if failed
end


makeLinkFile(1, [6, 8, 2, 4], 3)
makeLinkFile(2, [7, 1, 3, 5], 3)
makeLinkFile(3, [8, 2, 4, 6], 3)
makeLinkFile(4, [1, 3, 5, 7], 3)
makeLinkFile(5, [2, 4, 6, 8], 3)
makeLinkFile(6, [3, 5, 7, 1], 3)
makeLinkFile(7, [4, 6, 8, 2], 3)
makeLinkFile(8, [5, 7, 1, 3], 3)
